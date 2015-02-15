package emitter

import (
	"fmt"
)

func (f *Data) matchingIntLiteral(depth int, ctx *emitterContext, intNumber int) {

	f.PrintLabel(depth-1, fmt.Sprintf("//Matching %d literal", intNumber))

	f.printOffsetCheck(depth, ctx.patternCtx.prevEntryPoint, "")

	f.PrintLabel(depth, fmt.Sprintf("if (memMngr.vterms[fragmentOffset].tag != V_INT_NUM_TAG || "+
		"memMngr.vterms[fragmentOffset].intNum != %d)", intNumber))
	f.printFailBlock(depth, ctx.patternCtx.prevEntryPoint, true)

	f.PrintLabel(depth, "fragmentOffset++;")
}

func (f *Data) matchingCompLiteral(depth int, ctx *emitterContext, compSymbol string) {

	f.PrintLabel(depth-1, fmt.Sprintf("//Matching %q literal", compSymbol))

	f.printOffsetCheck(depth, ctx.patternCtx.prevEntryPoint, "")

	f.PrintLabel(depth, fmt.Sprintf("if (memMngr.vterms[fragmentOffset].tag != V_IDENT_TAG || "+
		"strcmp(memMngr.vterms[fragmentOffset].str, %q))", compSymbol))
	f.printFailBlock(depth, ctx.patternCtx.prevEntryPoint, true)

	f.PrintLabel(depth, "fragmentOffset++;")
}

func (f *Data) matchingStrLiteral(depth int, ctx *emitterContext, str string) {

	f.PrintLabel(depth-1, fmt.Sprintf("//Matching %q literal", str))

	f.printOffsetCheck(depth, ctx.patternCtx.prevEntryPoint, "")

	f.PrintLabel(depth, fmt.Sprintf("for (i = 0; i < %d; i++)", len(str)))
	f.PrintLabel(depth, "{")

	f.PrintLabel(depth+1, fmt.Sprintf("if (memMngr.vterms[fragmentOffset + i].tag != V_CHAR_TAG || "+
		"memMngr.vterms[fragmentOffset + i].ch != %q[i])", str))

	f.printFailBlock(depth+1, ctx.patternCtx.prevEntryPoint, true)

	f.PrintLabel(depth, "}")

	f.PrintLabel(depth, fmt.Sprintf("fragmentOffset += %d;", len(str)))
}