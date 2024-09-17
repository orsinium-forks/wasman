package wasm

import (
	"github.com/c0mm4nd/wasman/expr"
	"github.com/c0mm4nd/wasman/stacks"
)

// Frame is the context data of one instance
type Frame struct {
	PC         uint64
	Func       *wasmFunc
	Locals     []uint64
	LabelStack *stacks.Stack[*stacks.Label]
}

// instructions are basic wasm instructions
var instructions = [256]func(ins *Instance) error{
	expr.OpCodeUnreachable:       unreachable,
	expr.OpCodeNop:               nop,
	expr.OpCodeBlock:             block,
	expr.OpCodeLoop:              loop,
	expr.OpCodeIf:                ifOp,
	expr.OpCodeElse:              elseOp,
	expr.OpCodeEnd:               end,
	expr.OpCodeBr:                br,
	expr.OpCodeBrIf:              brIf,
	expr.OpCodeBrTable:           brTable,
	expr.OpCodeReturn:            nop,
	expr.OpCodeCall:              call,
	expr.OpCodeCallIndirect:      callIndirect,
	expr.OpCodeDrop:              drop,
	expr.OpCodeSelect:            selectOp,
	expr.OpCodeLocalGet:          getLocal,
	expr.OpCodeLocalSet:          setLocal,
	expr.OpCodeLocalTee:          teeLocal,
	expr.OpCodeGlobalGet:         getGlobal,
	expr.OpCodeGlobalSet:         setGlobal,
	expr.OpCodeI32Load:           i32Load,
	expr.OpCodeI64Load:           i64Load,
	expr.OpCodeF32Load:           f32Load,
	expr.OpCodeF64Load:           f64Load,
	expr.OpCodeI32Load8s:         i32Load8s,
	expr.OpCodeI32Load8u:         i32Load8u,
	expr.OpCodeI32Load16s:        i32Load16s,
	expr.OpCodeI32Load16u:        i32Load16u,
	expr.OpCodeI64Load8s:         i64Load8s,
	expr.OpCodeI64Load8u:         i64Load8u,
	expr.OpCodeI64Load16s:        i64Load16s,
	expr.OpCodeI64Load16u:        i64Load16u,
	expr.OpCodeI64Load32s:        i64Load32s,
	expr.OpCodeI64Load32u:        i64Load32u,
	expr.OpCodeI32Store:          i32Store,
	expr.OpCodeI64Store:          i64Store,
	expr.OpCodeF32Store:          f32Store,
	expr.OpCodeF64Store:          f64Store,
	expr.OpCodeI32Store8:         i32Store8,
	expr.OpCodeI32Store16:        i32Store16,
	expr.OpCodeI64Store8:         i64Store8,
	expr.OpCodeI64Store16:        i64Store16,
	expr.OpCodeI64Store32:        i64Store32,
	expr.OpCodeMemorySize:        memorySize,
	expr.OpCodeMemoryGrow:        memoryGrow,
	expr.OpCodeBulkMemory:        bulkMemory,
	expr.OpCodeI32Const:          i32Const,
	expr.OpCodeI64Const:          i64Const,
	expr.OpCodeF32Const:          f32Const,
	expr.OpCodeF64Const:          f64Const,
	expr.OpCodeI32Eqz:            i32eqz,
	expr.OpCodeI32Eq:             i32eq,
	expr.OpCodeI32Ne:             i32ne,
	expr.OpCodeI32LtS:            i32lts,
	expr.OpCodeI32LtU:            i32ltu,
	expr.OpCodeI32GtS:            i32gts,
	expr.OpCodeI32GtU:            i32gtu,
	expr.OpCodeI32LeS:            i32les,
	expr.OpCodeI32LeU:            i32leu,
	expr.OpCodeI32GeS:            i32ges,
	expr.OpCodeI32GeU:            i32geu,
	expr.OpCodeI64Eqz:            i64eqz,
	expr.OpCodeI64Eq:             i64eq,
	expr.OpCodeI64Ne:             i64ne,
	expr.OpCodeI64LtS:            i64lts,
	expr.OpCodeI64LtU:            i64ltu,
	expr.OpCodeI64GtS:            i64gts,
	expr.OpCodeI64GtU:            i64gtu,
	expr.OpCodeI64LeS:            i64les,
	expr.OpCodeI64LeU:            i64leu,
	expr.OpCodeI64GeS:            i64ges,
	expr.OpCodeI64GeU:            i64geu,
	expr.OpCodeF32Eq:             f32eq,
	expr.OpCodeF32Ne:             f32ne,
	expr.OpCodeF32Lt:             f32lt,
	expr.OpCodeF32Gt:             f32gt,
	expr.OpCodeF32Le:             f32le,
	expr.OpCodeF32Ge:             f32ge,
	expr.OpCodeF64Eq:             f64eq,
	expr.OpCodeF64Ne:             f64ne,
	expr.OpCodeF64Lt:             f64lt,
	expr.OpCodeF64Gt:             f64gt,
	expr.OpCodeF64Le:             f64le,
	expr.OpCodeF64Ge:             f64ge,
	expr.OpCodeI32Clz:            i32clz,
	expr.OpCodeI32Ctz:            i32ctz,
	expr.OpCodeI32PopCnt:         i32popcnt,
	expr.OpCodeI32Add:            i32add,
	expr.OpCodeI32Sub:            i32sub,
	expr.OpCodeI32Mul:            i32mul,
	expr.OpCodeI32DivS:           i32divs,
	expr.OpCodeI32DivU:           i32divu,
	expr.OpCodeI32RemS:           i32rems,
	expr.OpCodeI32RemU:           i32remu,
	expr.OpCodeI32And:            i32and,
	expr.OpCodeI32Or:             i32or,
	expr.OpCodeI32Xor:            i32xor,
	expr.OpCodeI32Shl:            i32shl,
	expr.OpCodeI32ShrS:           i32shrs,
	expr.OpCodeI32ShrU:           i32shru,
	expr.OpCodeI32RotL:           i32rotl,
	expr.OpCodeI32RotR:           i32rotr,
	expr.OpCodeI64Clz:            i64clz,
	expr.OpCodeI64Ctz:            i64ctz,
	expr.OpCodeI64PopCnt:         i64popcnt,
	expr.OpCodeI64Add:            i64add,
	expr.OpCodeI64Sub:            i64sub,
	expr.OpCodeI64Mul:            i64mul,
	expr.OpCodeI64DivS:           i64divs,
	expr.OpCodeI64DivU:           i64divu,
	expr.OpCodeI64RemS:           i64rems,
	expr.OpCodeI64RemU:           i64remu,
	expr.OpCodeI64And:            i64and,
	expr.OpCodeI64Or:             i64or,
	expr.OpCodeI64Xor:            i64xor,
	expr.OpCodeI64Shl:            i64shl,
	expr.OpCodeI64ShrS:           i64shrs,
	expr.OpCodeI64ShrU:           i64shru,
	expr.OpCodeI64RotL:           i64rotl,
	expr.OpCodeI64RotR:           i64rotr,
	expr.OpCodeF32Abs:            f32abs,
	expr.OpCodeF32Neg:            f32neg,
	expr.OpCodeF32Ceil:           f32ceil,
	expr.OpCodeF32Floor:          f32floor,
	expr.OpCodeF32Trunc:          f32trunc,
	expr.OpCodeF32Nearest:        f32nearest,
	expr.OpCodeF32Sqrt:           f32sqrt,
	expr.OpCodeF32Add:            f32add,
	expr.OpCodeF32Sub:            f32sub,
	expr.OpCodeF32Mul:            f32mul,
	expr.OpCodeF32Div:            f32div,
	expr.OpCodeF32Min:            f32min,
	expr.OpCodeF32Max:            f32max,
	expr.OpCodeF32CopySign:       f32copysign,
	expr.OpCodeF64Abs:            f64abs,
	expr.OpCodeF64Neg:            f64neg,
	expr.OpCodeF64Ceil:           f64ceil,
	expr.OpCodeF64Floor:          f64floor,
	expr.OpCodeF64Trunc:          f64trunc,
	expr.OpCodeF64Nearest:        f64nearest,
	expr.OpCodeF64Sqrt:           f64sqrt,
	expr.OpCodeF64Add:            f64add,
	expr.OpCodeF64Sub:            f64sub,
	expr.OpCodeF64Mul:            f64mul,
	expr.OpCodeF64Div:            f64div,
	expr.OpCodeF64Min:            f64min,
	expr.OpCodeF64Max:            f64max,
	expr.OpCodeF64CopySign:       f64copysign,
	expr.OpCodeI32WrapI64:        i32wrapi64,
	expr.OpCodeI32TruncF32S:      i32truncf32s,
	expr.OpCodeI32TruncF32U:      i32truncf32u,
	expr.OpCodeI32truncF64S:      i32truncf64s,
	expr.OpCodeI32truncF64U:      i32truncf64u,
	expr.OpCodeI64ExtendI32S:     i64extendi32s,
	expr.OpCodeI64ExtendI32U:     i64extendi32u,
	expr.OpCodeI64TruncF32S:      i64truncf32s,
	expr.OpCodeI64TruncF32U:      i64truncf32u,
	expr.OpCodeI64TruncF64S:      i64truncf64s,
	expr.OpCodeI64TruncF64U:      i64truncf64u,
	expr.OpCodeF32ConvertI32S:    f32converti32s,
	expr.OpCodeF32ConvertI32U:    f32converti32u,
	expr.OpCodeF32ConvertI64S:    f32converti64s,
	expr.OpCodeF32ConvertI64U:    f32converti64u,
	expr.OpCodeF32DemoteF64:      f32demotef64,
	expr.OpCodeF64ConvertI32S:    f64converti32s,
	expr.OpCodeF64ConvertI32U:    f64converti32u,
	expr.OpCodeF64ConvertI64S:    f64converti64s,
	expr.OpCodeF64ConvertI64U:    f64converti64u,
	expr.OpCodeF64PromoteF32:     f64promotef32,
	expr.OpCodeI32ReinterpretF32: nop,
	expr.OpCodeI64ReinterpretF64: nop,
	expr.OpCodeF32ReinterpretI32: nop,
	expr.OpCodeF64ReinterpretI64: nop,
}
