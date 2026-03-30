import { ref, reactive, computed } from 'vue'
import JSZip from 'jszip'

const API_BASE = '/api'

// Prefix configs (loaded from backend, with fallback)
const defaultPrefixes = [
    { prefix: 'bicom_', label: 'bicom_greeting', targetDb: -12, description: 'Bicom Greeting' },
    { prefix: 'aa_', label: 'auto_attendant', targetDb: -6, description: 'Auto Attendant' },
    { prefix: 'mbx_', label: 'mailbox_greeting', targetDb: -6, description: 'Mailbox Greeting' },
    { prefix: 'moh_', label: 'hold_music', targetDb: -20, description: 'Hold Music' },
]

export function useConverter() {
    // State
    const files = ref([])
    const showCelebration = ref(false)

    const options = reactive({
        format: 'wav-pcm',
        normalize: true,
        targetDb: -6,
        bandpass: false,
        bandpassLow: 300,
        bandpassHigh: 3400,
    })

    const selectedPreset = ref('global')

    const presetOptions = [
        { id: 'global', label: 'Global (Custom)', description: 'Use global settings' },
        { id: 'bicom_', label: 'Bicom Greeting', format: 'wav-pcm', targetDb: -12, description: '8kHz WAV, -12 dBI peak' },
        { id: 'aa_', label: 'Auto Attendant', format: 'wav-pcm', targetDb: -6, description: '8kHz WAV, -6 dB peak' },
        { id: 'mbx_', label: 'Mailbox', format: 'wav-pcm', targetDb: -6, description: '8kHz WAV, -6 dB peak' },
        { id: 'moh_', label: 'Hold Music', format: 'wav-pcm', targetDb: -20, description: '8kHz WAV, -20 dB peak' },
    ]

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

    // Get effective preset for a file (inline override > filename match > global)
    function effectivePreset(fileEntry) {
        if (fileEntry.presetOverride) {
            return prefixes.value.find(p => p.prefix === fileEntry.presetOverride) || null
        }
        if (fileEntry.prefix && fileEntry.prefix.prefix) {
            return fileEntry.prefix
        }
        return null  // global settings will apply
    }

    // Apply a preset to global options
    function setPreset(presetId) {
        selectedPreset.value = presetId
        if (presetId === 'global') return

        const preset = presetOptions.find(p => p.id === presetId)
        if (preset) {
            if (preset.targetDb !== undefined) {
                options.targetDb = preset.targetDb
                options.normalize = true
            }
            if (preset.format) {
                options.format = preset.format
            }
        }
    }

    // Add files from input or drop
    function addFiles(fileList) {
        const newFiles = Array.from(fileList).map(f => ({
            id: crypto.randomUUID(),
            file: f,
            name: f.name,
            size: f.size,
            prefix: detectPrefix(f.name),
            presetOverride: null,  // null | 'bicom_' | 'aa_' | 'mbx_' | 'moh_'
            status: 'pending',  // pending | converting | done | error
            progress: 0,
            error: '',
            resultBlob: null,
            resultFilename: '',
            audioStats: null,  // { inputLoudness, inputPeak, inputLRA, outputLoudness, outputPeak, outputLRA }
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

    function setFilePreset(id, presetOverride) {
        const file = files.value.find(f => f.id === id)
        if (file) {
            file.presetOverride = presetOverride || null
        }
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
        const effPreset = effectivePreset(fileEntry)
        const effTargetDb = effPreset ? effPreset.targetDb : options.targetDb
        const effFormat = effPreset && effPreset.format ? effPreset.format : options.format

        const formData = new FormData()
        formData.append('file', fileEntry.file)
        formData.append('format', effFormat)
        formData.append('normalize', options.normalize ? 'true' : 'false')
        formData.append('target_db', effTargetDb.toString())
        formData.append('bandpass', options.bandpass ? 'true' : 'false')
        formData.append('bandpass_low', options.bandpassLow.toString())
        formData.append('bandpass_high', options.bandpassHigh.toString())

        try {
            const { blob, headers } = await uploadWithProgress(
                `${API_BASE}/convert`,
                formData,
                (progress) => {
                    fileEntry.progress = progress
                }
            )

            fileEntry.status = 'done'
            fileEntry.resultBlob = blob

            // Capture audio stats from response headers
            fileEntry.audioStats = {
                inputLoudness: parseFloat(headers.get('X-Input-Loudness')) || null,
                inputPeak: parseFloat(headers.get('X-Input-Peak')) || null,
                inputLRA: parseFloat(headers.get('X-Input-LRA')) || null,
                outputLoudness: parseFloat(headers.get('X-Output-Loudness')) || null,
                outputPeak: parseFloat(headers.get('X-Output-Peak')) || null,
                outputLRA: parseFloat(headers.get('X-Output-LRA')) || null,
                targetDb: parseFloat(headers.get('X-Normalization-DB')) || null,
            }

            // Determine output filename: strip prefix, add timestamp
            let baseName = fileEntry.name.replace(/\.[^.]+$/, '')
            // Strip the effective prefix (either auto-detected or manual override)
            const prefixToStrip = fileEntry.presetOverride || (fileEntry.prefix && fileEntry.prefix.prefix) || ''
            if (prefixToStrip) {
                const pfx = prefixToStrip
                if (baseName.toLowerCase().startsWith(pfx)) {
                    baseName = baseName.substring(pfx.length)
                }
            }
            // Append _c-TIMESTAMP in browser-local time
            const now = new Date()
            const ts = now.getFullYear().toString()
                + String(now.getMonth() + 1).padStart(2, '0')
                + String(now.getDate()).padStart(2, '0')
                + '-'
                + String(now.getHours()).padStart(2, '0')
                + String(now.getMinutes()).padStart(2, '0')
                + String(now.getSeconds()).padStart(2, '0')
            fileEntry.resultFilename = baseName + '_c-' + ts + '.wav'

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
                    // Build a headers-like object from XHR
                    const headersObj = {
                        get(name) {
                            return xhr.getResponseHeader(name)
                        }
                    }
                    resolve({ blob: xhr.response, headers: headersObj })
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
        triggerCelebration()
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
        const now = new Date()
        const ts = now.getFullYear().toString()
            + String(now.getMonth() + 1).padStart(2, '0')
            + String(now.getDate()).padStart(2, '0')
            + '-'
            + String(now.getHours()).padStart(2, '0')
            + String(now.getMinutes()).padStart(2, '0')
            + String(now.getSeconds()).padStart(2, '0')
        a.download = 'converted_audio_' + ts + '.zip'
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(url)
        triggerCelebration()
    }

    // Easter egg celebration
    function triggerCelebration() {
        if (document.body.classList.contains('easter-egg')) {
            showCelebration.value = true
        }
    }

    function dismissCelebration() {
        showCelebration.value = false
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
        selectedPreset,
        presetOptions,
        isProcessing,
        hasFiles,
        hasPendingFiles,
        hasDoneFiles,
        pendingFiles,
        doneFiles,
        errorFiles,
        overallStatus,
        overallProgress,
        showCelebration,
        detectPrefix,
        effectivePreset,
        setPreset,
        setFilePreset,
        addFiles,
        removeFile,
        clearFiles,
        convertAll,
        downloadFile,
        downloadAllAsZip,
        dismissCelebration,
        loadFormats,
        loadPrefixes,
    }
}
