import { ref, reactive, computed } from 'vue'
import JSZip from 'jszip'

const API_BASE = '/api'

// Prefix configs (loaded from backend, with fallback)
const defaultPrefixes = [
    { prefix: 'aa_', label: 'auto_attendant', targetDb: -6, description: 'Auto Attendant' },
    { prefix: 'mbx_', label: 'mailbox_greeting', targetDb: -6, description: 'Mailbox Greeting' },
    { prefix: 'moh_', label: 'hold_music', targetDb: -20, description: 'Hold Music' },
]

export function useConverter() {
    // State
    const files = ref([])

    const options = reactive({
        format: 'wav-pcm',
        normalize: true,
        bandpass: false,
        bandpassLow: 300,
        bandpassHigh: 3400,
    })

    const formats = ref([
        { id: 'wav-pcm', label: 'Standard WAV (8kHz, 16-bit PCM)', sampleRate: 8000 },
        { id: 'wav-ulaw', label: 'G.711 µ-law WAV (8kHz)', sampleRate: 8000 },
        { id: 'wav-alaw', label: 'G.711 A-law WAV (8kHz)', sampleRate: 8000 },
        { id: 'g722', label: 'G.722 (16kHz)', sampleRate: 16000 },
    ])

    const prefixes = ref(defaultPrefixes)

    // Computed
    const hasFiles = computed(() => files.value.length > 0)
    const isProcessing = computed(() => files.value.some(f => f.status === 'converting'))

    const pendingFiles = computed(() => files.value.filter(f => f.status === 'pending'))
    const doneFiles = computed(() => files.value.filter(f => f.status === 'done'))
    const errorFiles = computed(() => files.value.filter(f => f.status === 'error'))
    const hasPendingFiles = computed(() => pendingFiles.value.length > 0)
    const hasDoneFiles = computed(() => doneFiles.value.length > 0)

    // Overall status for the progress tracker
    const overallStatus = computed(() => {
        if (files.value.length === 0) return ''
        if (files.value.some(f => f.status === 'converting')) return 'converting'
        if (files.value.every(f => f.status === 'done')) return 'done'
        if (files.value.some(f => f.status === 'done') && !files.value.some(f => f.status === 'converting' || f.status === 'pending')) {
            // Some done, some error, none pending/converting
            return files.value.some(f => f.status === 'error') ? 'partial' : 'done'
        }
        if (files.value.every(f => f.status === 'error')) return 'error'
        return ''
    })

    const overallProgress = computed(() => {
        const total = files.value.length
        if (total === 0) return 0
        const completed = files.value.filter(f => f.status === 'done' || f.status === 'error').length
        const inProgress = files.value.filter(f => f.status === 'converting')
        const progressFromConverting = inProgress.reduce((sum, f) => sum + f.progress, 0) / total
        return Math.round(((completed / total) * 100) + progressFromConverting)
    })

    // Detect prefix for a filename
    function detectPrefix(filename) {
        const lower = filename.toLowerCase()
        for (const p of prefixes.value) {
            if (lower.startsWith(p.prefix)) {
                return p
            }
        }
        return { prefix: '', label: 'unknown', targetDb: -6, description: 'Standard' }
    }

    // Add files from input or drop
    function addFiles(fileList) {
        const newFiles = Array.from(fileList).map(f => ({
            id: crypto.randomUUID(),
            file: f,
            name: f.name,
            size: f.size,
            prefix: detectPrefix(f.name),
            status: 'pending',  // pending | converting | done | error
            progress: 0,
            error: '',
            resultBlob: null,
            resultFilename: '',
        }))
        files.value = [...files.value, ...newFiles]
    }

    function removeFile(id) {
        const file = files.value.find(f => f.id === id)
        if (file && file.resultBlob) {
            URL.revokeObjectURL(file._objectUrl)
        }
        files.value = files.value.filter(f => f.id !== id)
    }

    function clearFiles() {
        // Clean up all blob URLs
        for (const f of files.value) {
            if (f._objectUrl) {
                URL.revokeObjectURL(f._objectUrl)
            }
        }
        files.value = []
    }

    // Convert all pending files independently
    async function convertAll() {
        const toConvert = files.value.filter(f => f.status === 'pending' || f.status === 'error')
        if (toConvert.length === 0) return

        // Reset error files back to pending
        toConvert.forEach(f => {
            f.status = 'converting'
            f.progress = 0
            f.error = ''
        })

        // Convert each file independently (concurrent, max 3 at a time)
        const concurrency = 3
        const queue = [...toConvert]

        async function processNext() {
            while (queue.length > 0) {
                const file = queue.shift()
                await convertSingleFile(file)
            }
        }

        const workers = Array.from({ length: Math.min(concurrency, queue.length) }, () => processNext())
        await Promise.all(workers)
    }

    // Convert a single file via /api/convert
    async function convertSingleFile(fileEntry) {
        const formData = new FormData()
        formData.append('file', fileEntry.file)
        formData.append('format', options.format)
        formData.append('normalize', options.normalize ? 'true' : 'false')
        formData.append('bandpass', options.bandpass ? 'true' : 'false')
        formData.append('bandpass_low', options.bandpassLow.toString())
        formData.append('bandpass_high', options.bandpassHigh.toString())

        try {
            const blob = await uploadWithProgress(
                `${API_BASE}/convert`,
                formData,
                (progress) => {
                    fileEntry.progress = progress
                }
            )

            fileEntry.status = 'done'
            fileEntry.resultBlob = blob

            // Determine output filename
            const baseName = fileEntry.name.replace(/\.[^.]+$/, '')
            fileEntry.resultFilename = baseName + '_converted.wav'

        } catch (err) {
            fileEntry.status = 'error'
            fileEntry.error = err.message || 'Conversion failed'
        }
    }

    function uploadWithProgress(url, formData, onProgress) {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest()
            xhr.open('POST', url)

            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const pct = Math.round((e.loaded / e.total) * 100)
                    onProgress(pct)
                }
            })

            xhr.addEventListener('load', () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    resolve(xhr.response)
                } else {
                    reject(new Error(xhr.responseText || `Server error: ${xhr.status}`))
                }
            })

            xhr.addEventListener('error', () => reject(new Error('Network error')))
            xhr.addEventListener('abort', () => reject(new Error('Upload aborted')))

            xhr.responseType = 'blob'
            xhr.send(formData)
        })
    }

    // Download a single converted file
    function downloadFile(id) {
        const file = files.value.find(f => f.id === id)
        if (!file || !file.resultBlob) return

        if (!file._objectUrl) {
            file._objectUrl = URL.createObjectURL(file.resultBlob)
        }
        const a = document.createElement('a')
        a.href = file._objectUrl
        a.download = file.resultFilename
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
    }

    // Download all converted files as a ZIP
    async function downloadAllAsZip() {
        const completed = files.value.filter(f => f.status === 'done' && f.resultBlob)
        if (completed.length === 0) return

        const zip = new JSZip()
        for (const f of completed) {
            zip.file(f.resultFilename, f.resultBlob)
        }

        const zipBlob = await zip.generateAsync({ type: 'blob' })
        const url = URL.createObjectURL(zipBlob)
        const a = document.createElement('a')
        a.href = url
        a.download = 'converted_audio.zip'
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(url)
    }

    // Load formats from backend
    async function loadFormats() {
        try {
            const res = await fetch(`${API_BASE}/formats`)
            if (res.ok) {
                formats.value = await res.json()
            }
        } catch {
            // Use defaults
        }
    }

    // Load prefixes from backend
    async function loadPrefixes() {
        try {
            const res = await fetch(`${API_BASE}/prefixes`)
            if (res.ok) {
                prefixes.value = await res.json()
            }
        } catch {
            // Use defaults
        }
    }

    return {
        files,
        options,
        formats,
        prefixes,
        isProcessing,
        hasFiles,
        hasPendingFiles,
        hasDoneFiles,
        pendingFiles,
        doneFiles,
        errorFiles,
        overallStatus,
        overallProgress,
        detectPrefix,
        addFiles,
        removeFile,
        clearFiles,
        convertAll,
        downloadFile,
        downloadAllAsZip,
        loadFormats,
        loadPrefixes,
    }
}
