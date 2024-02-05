package wasman

import (
	"errors"
	"fmt"

	"github.com/c0mm4nd/wasman/config"
	"github.com/c0mm4nd/wasman/wasm"

	"github.com/c0mm4nd/wasman/segments"
	"github.com/c0mm4nd/wasman/types"
)

// errors on linking modules
var (
	ErrInvalidSign = errors.New("invalid signature")
)

// Linker is a helper to instantiate new modules
type Linker struct {
	config.LinkerConfig

	Modules map[string]*Module // the built-in modules which acts as externs when instantiating coming main module
}

// NewLinker creates a new Linker
func NewLinker(config config.LinkerConfig) *Linker {
	return &Linker{
		LinkerConfig: config,
		Modules:      map[string]*Module{},
	}
}

// NewLinkerWithModuleMap creates a new Linker with the built-in modules
func NewLinkerWithModuleMap(config config.LinkerConfig, in map[string]*Module) *Linker {
	return &Linker{
		LinkerConfig: config,
		Modules:      in,
	}
}

// Define put the module on its namespace
func (l *Linker) Define(modName string, mod *Module) {
	l.Modules[modName] = mod
}

// AdvancedFunc is a advanced host func comparing to normal go host func
// Dev will be able to handle the pre/post-call process of the func and manipulate
// the Instance's fields like memory
//
// e.g. when we wanna add toll after calling the host func f
//
//	func ExampleFuncGenerator_addToll() {
//		var linker = wasman.NewLinker()
//		var f = func() {fmt.Println("wasm")}
//
//		var af = wasman.AdvancedFunc(func(ins *wasman.Instance) any {
//			return func() {
//				f()
//				ins.AddGas(11)
//			}
//		})
//		linker.DefineAdvancedFunc("env", "add_gas", af)
//	}
//
// e.g. when we wanna manipulate memory
//
//	func ExampleFuncGenerator_addToll() {
//		var linker = wasman.NewLinker()
//
//		var af = wasman.AdvancedFunc(func(ins *wasman.Instance) any {
//			return func(ptr uint32, length uint32) {
//				msg := ins.Memory[int(ptr), int(ptr+uint32)]
//				fmt.Println(b)
//			}
//		})
//
//		linker.DefineAdvancedFunc("env", "print_msg", af)
//	}
type AdvancedFunc func(ins *Instance) any

func DefineFunc10[A any](l *Linker, modName, funcName string, f func(A)) error {
	sig, err := getSignatureN0([]any{*new(A)})
	if err != nil {
		return ErrInvalidSign
	}
	return l.defineFunc(modName, funcName, sig, f)
}

func DefineFunc11[A, Z any](l *Linker, modName, funcName string, f func(A) Z) error {
	sig, err := getSignatureN1([]any{*new(A)}, *new(Z))
	if err != nil {
		return ErrInvalidSign
	}
	return l.defineFunc(modName, funcName, sig, f)
}

// DefineFunc puts a simple go style func into Linker's modules.
// This f should be a simply func which doesnt handle ins's fields.
func (l *Linker) defineFunc(modName, funcName string, sig *types.FuncType, f any) error {
	fn := func(ins *Instance) any {
		return f
	}
	mod, exists := l.Modules[modName]
	if !exists {
		mod = &Module{IndexSpace: new(wasm.IndexSpace), ExportSection: map[string]*segments.ExportSegment{}}
		l.Modules[modName] = mod
	}

	if l.DisableShadowing && mod.ExportSection[funcName] != nil {
		return config.ErrShadowing
	}

	mod.ExportSection[funcName] = &segments.ExportSegment{
		Name: funcName,
		Desc: &segments.ExportDesc{
			Kind:  segments.KindFunction,
			Index: uint32(len(mod.IndexSpace.Functions)),
		},
	}

	mod.IndexSpace.Functions = append(mod.IndexSpace.Functions, &wasm.HostFunc{
		Generator: fn,
		Signature: sig,
	})

	return nil
}

func DefineGlobal[T any](l *Linker, modName, globalName string, global T) error {
	ty, err := getTypeOf(*new(T))
	if err != nil {
		return err
	}
	return l.defineGlobal(modName, globalName, ty, global)
}

// DefineGlobal will defined an external global for the main module
func (l *Linker) defineGlobal(modName, globalName string, ty types.ValueType, global any) error {
	mod, exists := l.Modules[modName]
	if !exists {
		mod = &Module{IndexSpace: new(wasm.IndexSpace), ExportSection: map[string]*segments.ExportSegment{}}
		l.Modules[modName] = mod
	}

	if l.DisableShadowing && mod.ExportSection[globalName] != nil {
		return config.ErrShadowing
	}

	mod.ExportSection[globalName] = &segments.ExportSegment{
		Name: globalName,
		Desc: &segments.ExportDesc{
			Kind:  segments.KindGlobal,
			Index: uint32(len(mod.IndexSpace.Globals)),
		},
	}

	mod.IndexSpace.Globals = append(mod.IndexSpace.Globals, &wasm.Global{
		GlobalType: &types.GlobalType{
			ValType: ty,
			Mutable: true,
		},
		Val: global,
	})

	return nil
}

// DefineTable will defined an external table for the main module
func (l *Linker) DefineTable(modName, tableName string, table []*uint32) error {
	mod, exists := l.Modules[modName]
	if !exists {
		mod = &Module{IndexSpace: new(wasm.IndexSpace), ExportSection: map[string]*segments.ExportSegment{}}
		l.Modules[modName] = mod
	}

	if l.DisableShadowing && mod.ExportSection[tableName] != nil {
		return config.ErrShadowing
	}

	mod.ExportSection[tableName] = &segments.ExportSegment{
		Name: tableName,
		Desc: &segments.ExportDesc{
			Kind:  segments.KindTable,
			Index: uint32(len(mod.IndexSpace.Tables)),
		},
	}

	mod.IndexSpace.Tables = append(mod.IndexSpace.Tables, &wasm.Table{
		TableType: *mod.TableSection[0],
		Value:     table,
	})

	return nil
}

// DefineMemory will defined an external memory for the main module
func (l *Linker) DefineMemory(modName, memName string, mem []byte) error {
	mod, exists := l.Modules[modName]
	if !exists {
		mod = &Module{IndexSpace: new(wasm.IndexSpace), ExportSection: map[string]*segments.ExportSegment{}}
		l.Modules[modName] = mod
	}

	if l.DisableShadowing && mod.ExportSection[memName] != nil {
		return config.ErrShadowing
	}

	mod.ExportSection[memName] = &segments.ExportSegment{
		Name: memName,
		Desc: &segments.ExportDesc{
			Kind:  segments.KindMem,
			Index: uint32(len(mod.IndexSpace.Memories)),
		},
	}

	mod.IndexSpace.Memories = append(mod.IndexSpace.Memories, &wasm.Memory{
		MemoryType: *mod.MemorySection[0],
		Value:      mem,
	})

	return nil
}

// Instantiate will instantiate a Module into an runnable Instance
func (l *Linker) Instantiate(mainModule *Module) (*Instance, error) {
	return NewInstance(mainModule, l.Modules)
}

func getSignatureN0(in []any) (*types.FuncType, error) {
	var err error
	ins, err := getTypesOf(in...)
	if err != nil {
		return nil, err
	}
	return &types.FuncType{InputTypes: ins, ReturnTypes: []types.ValueType{}}, nil
}

func getSignatureN1(in []any, out any) (*types.FuncType, error) {
	var err error
	ins, err := getTypesOf(in...)
	if err != nil {
		return nil, err
	}
	outs, err := getTypesOf(out)
	if err != nil {
		return nil, err
	}
	return &types.FuncType{InputTypes: ins, ReturnTypes: outs}, nil
}

func getTypesOf(defaults ...any) ([]types.ValueType, error) {
	var err error
	types := make([]types.ValueType, len(defaults))
	for i, def := range defaults {
		types[i], err = getTypeOf(def)
		if err != nil {
			return nil, err
		}
	}
	return types, nil
}

// getTypeOf converts the go type into wasm val type
func getTypeOf(def any) (types.ValueType, error) {
	switch def.(type) {
	case float64:
		return types.ValueTypeF64, nil
	case float32:
		return types.ValueTypeF32, nil
	case int32, uint32, int, int16, int8, bool:
		return types.ValueTypeI32, nil
	case int64, uint64, uintptr, uint:
		return types.ValueTypeI64, nil
	default:
		return 0x00, fmt.Errorf("invalid type: %T", def)
	}

}
