import { ref, reactive, computed } from 'vue'

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
    const isUploading = ref(false)
    const isConverting = ref(false)
    const uploadProgress = ref(0)
    const conversionStatus = ref('') // idle, uploading, converting, done, error
    const resultUrl = ref(null)
    const resultFilename = ref('')
    const error = ref('')

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
    const isBulk = computed(() => {
        return files.value.length > 1 || (files.value.length === 1 && files.value[0].name.toLowerCase().endsWith('.zip'))
    })
    const isProcessing = computed(() => isUploading.value || isConverting.value)

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
        error.value = ''
        const newFiles = Array.from(fileList).map(f => ({
            file: f,
            name: f.name,
            size: f.size,
            prefix: detectPrefix(f.name),
        }))
        files.value = [...files.value, ...newFiles]
    }

    function removeFile(index) {
        files.value.splice(index, 1)
    }

    function clearFiles() {
        files.value = []
        resultUrl.value = null
        resultFilename.value = ''
        error.value = ''
        conversionStatus.value = ''
        uploadProgress.value = 0
    }

    // Upload and convert
    async function uploadAndConvert() {
        if (!hasFiles.value) return

        error.value = ''
        isUploading.value = true
        conversionStatus.value = 'uploading'
        uploadProgress.value = 0

        // Clean up previous result URL
        if (resultUrl.value) {
            URL.revokeObjectURL(resultUrl.value)
            resultUrl.value = null
        }

        const formData = new FormData()
        formData.append('format', options.format)
        formData.append('normalize', options.normalize ? 'true' : 'false')
        formData.append('bandpass', options.bandpass ? 'true' : 'false')
        formData.append('bandpass_low', options.bandpassLow.toString())
        formData.append('bandpass_high', options.bandpassHigh.toString())

        let endpoint = `${API_BASE}/convert`

        if (isBulk.value && !files.value[0].name.toLowerCase().endsWith('.zip')) {
            // Multiple files: create a ZIP on the client side
            // For now, we only support single file or ZIP upload
            // If multiple files, we convert them one-by-one
            // TODO: client-side ZIP creation for multi-file uploads
            // For now, convert first file only
            formData.append('file', files.value[0].file)
        } else if (files.value[0].name.toLowerCase().endsWith('.zip')) {
            endpoint = `${API_BASE}/convert/bulk`
            formData.append('file', files.value[0].file)
        } else {
            formData.append('file', files.value[0].file)
        }

        try {
            const blob = await uploadWithProgress(endpoint, formData, (progress) => {
                uploadProgress.value = progress
                if (progress >= 100) {
                    isUploading.value = false
                    isConverting.value = true
                    conversionStatus.value = 'converting'
                }
            })

            isConverting.value = false
            conversionStatus.value = 'done'

            resultUrl.value = URL.createObjectURL(blob)

            // Determine filename from content-disposition or generate one
            const isZip = files.value[0].name.toLowerCase().endsWith('.zip')
            if (isZip) {
                resultFilename.value = files.value[0].name.replace('.zip', '_converted.zip')
            } else {
                const baseName = files.value[0].name.replace(/\.[^.]+$/, '')
                const formatObj = formats.value.find(f => f.id === options.format)
                resultFilename.value = baseName + '_converted' + (formatObj ? '.wav' : '.wav')
            }
        } catch (err) {
            isUploading.value = false
            isConverting.value = false
            conversionStatus.value = 'error'
            error.value = err.message || 'Conversion failed'
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

    function downloadResult() {
        if (!resultUrl.value) return
        const a = document.createElement('a')
        a.href = resultUrl.value
        a.download = resultFilename.value
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
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

    return {
        files,
        options,
        formats,
        prefixes,
        isUploading,
        isConverting,
        isProcessing,
        uploadProgress,
        conversionStatus,
        resultUrl,
        resultFilename,
        error,
        hasFiles,
        isBulk,
        detectPrefix,
        addFiles,
        removeFile,
        clearFiles,
        uploadAndConvert,
        downloadResult,
        loadFormats,
    }
}
