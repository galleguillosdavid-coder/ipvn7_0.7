//go:build js && wasm

package main

import (
	"syscall/js"

	"ipvn7/pkg/wasm"
)

func main() {
	engine := wasm.NewWasmEngine()
	root := js.Global()

	ipvn7Obj := root.Get("Object").New()

	// 1. version()
	ipvn7Obj.Set("version", js.FuncOf(func(this js.Value, args []js.Value) any {
		return engine.Version()
	}))

	// 2. generateIdentity()
	ipvn7Obj.Set("generateIdentity", js.FuncOf(func(this js.Value, args []js.Value) any {
		id, err := engine.GenerateIdentity()
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{
			"did":         id.DID,
			"public_key":  id.PublicKey,
			"private_key": id.PrivateKey,
		}
	}))

	// 3. sign(privHex, messageStr)
	ipvn7Obj.Set("sign", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 2 {
			return map[string]any{"error": "se requieren privHex y message"}
		}
		privHex := args[0].String()
		msg := []byte(args[1].String())
		sig, err := engine.SignMessage(privHex, msg)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"signature": sig}
	}))

	// 4. verify(pubHex, messageStr, sigHex)
	ipvn7Obj.Set("verify", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 3 {
			return false
		}
		pubHex := args[0].String()
		msg := []byte(args[1].String())
		sigHex := args[2].String()
		return engine.VerifySignature(pubHex, msg, sigHex)
	}))

	// 5. computePoW(did, seq, difficulty)
	ipvn7Obj.Set("computePoW", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 2 {
			return map[string]any{"error": "se requieren did y seq"}
		}
		did := args[0].String()
		seq := uint64(args[1].Int())
		diff := 8
		if len(args) >= 3 {
			diff = args[2].Int()
		}
		nonce, err := engine.ComputeProofOfWork(did, seq, diff)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"nonce": nonce}
	}))

	// 6. validateFrame(hexStr)
	ipvn7Obj.Set("validateFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return false
		}
		ok, err := engine.ValidatePacketFrame(args[0].String())
		return ok && err == nil
	}))

	root.Set("ipvn7", ipvn7Obj)

	// Mantener residente el runtime WASM en el navegador
	select {}
}
