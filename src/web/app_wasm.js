// app_wasm.js - Puente universal WebAssembly para IPVN7 en el navegador
// Expone window.IPVN7Wasm con APIs criptográficas y de red soberana sin dependencias.

(function() {
    'use strict';

    class IPVN7WasmBridge {
        constructor() {
            this.ready = false;
            this.initPromise = null;
        }

        async init(wasmUrl = '/ipvn7.wasm') {
            if (this.initPromise) return this.initPromise;

            this.initPromise = (async () => {
                if (typeof Go === 'undefined') {
                    throw new Error('IPVN7Wasm: wasm_exec.js debe ser cargado antes de inicializar WASM');
                }
                const go = new Go();
                let response = await fetch(wasmUrl);
                if (!response.ok) {
                    throw new Error(`IPVN7Wasm: no se pudo cargar ${wasmUrl} (${response.status})`);
                }
                const bytes = await response.arrayBuffer();
                const result = await WebAssembly.instantiate(bytes, go.importObject);
                go.run(result.instance);
                this.ready = true;
                console.log('[IPVN7] WebAssembly Core inicializado exitosamente. Versión:', window.ipvn7.version());
                return true;
            })();

            return this.initPromise;
        }

        isReady() {
            return this.ready && typeof window.ipvn7 !== 'undefined';
        }

        async generateIdentity() {
            await this.init();
            return window.ipvn7.generateIdentity();
        }

        async sign(privHex, message) {
            await this.init();
            return window.ipvn7.sign(privHex, message);
        }

        async verify(pubHex, message, sigHex) {
            await this.init();
            return window.ipvn7.verify(pubHex, message, sigHex);
        }

        async computePoW(did, seq, difficulty = 8) {
            await this.init();
            return window.ipvn7.computePoW(did, seq, difficulty);
        }

        async validateFrame(hexStr) {
            await this.init();
            return window.ipvn7.validateFrame(hexStr);
        }

        version() {
            if (!this.isReady()) return 'WASM no inicializado';
            return window.ipvn7.version();
        }
    }

    window.IPVN7Wasm = new IPVN7WasmBridge();
})();
