package emitter

import (
	"fmt"
)

import (
	"BMSTU-Refal-Compiler/syntax"
	"BMSTU-Refal-Compiler/tokens"
)

func (f *Data) matchingPattern(depth int, ctx *emitterContext, terms []*syntax.Term) {

	f.PrintLabel(depth, fmt.Sprintf("//Sentence: %d, Pattern: %d", ctx.sentenceNumber, ctx.patternNumber))
	f.PrintLabel(depth, fmt.Sprintf("case %d:", ctx.entryPoint))
	f.PrintLabel(depth, fmt.Sprintf("{"))

	f.checkAndAssemblyChain(depth+1, ctx.patternNumber)

	f.PrintLabel(depth+1, "fragmentOffset = currFrag->offset;")
	f.PrintLabel(depth+1, fmt.Sprintf("stretchingVarNumber = env->stretchVarsNumber[%d];", ctx.entryPoint))

	f.PrintLabel(depth+1, "while (stretchingVarNumber >= 0)")
	f.PrintLabel(depth+1, "{")

	if len(terms) == 0 {
		f.processEmptyPattern(depth+1, ctx)
	} else {
		f.processPattern(depth+1, ctx, terms)
	}

	f.PrintLabel(depth+1, "} // Pattern while\n")

	f.processPatternFail(depth+1, ctx)

	ctx.prevEntryPoint = ctx.entryPoint
	ctx.entryPoint++
	ctx.patternNumber++
}

func (f *Data) processEmptyPattern(depth int, ctx *emitterContext) {
	f.PrintLabel(depth+1, "if (currFrag->length > 0)")
	f.printFailBlock(depth+1, -1, false)
	f.PrintLabel(depth+1, "break;")
}

func (f *Data) processPattern(depth int, ctx *emitterContext, terms []*syntax.Term) {

	f.PrintLabel(depth+1, "//From what stretchable variable start?")
	f.PrintLabel(depth+1, "switch (stretchingVarNumber)")
	f.PrintLabel(depth+1, "{")

	ctx.patternCtx.entryPoint = 0
	ctx.patternCtx.prevEntryPoint = -1

	f.printFirstCase(depth, ctx, terms[0])

	f.matchingTerm(depth+2, ctx, terms)

	f.PrintLabel(depth+1, "} //pattern switch\n")

	f.PrintLabel(depth+1, "if (!stretching)")
	f.PrintLabel(depth+1, "{")
	f.PrintLabel(depth+2, "if (fragmentOffset - currFrag->offset < currFrag->length)")
	f.printFailBlock(depth+2, ctx.patternCtx.prevEntryPoint, false)
	f.PrintLabel(depth+2, "else")
	f.PrintLabel(depth+3, "break; // Success!")
	f.PrintLabel(depth+1, "}")
}

func (f *Data) printFirstCase(depth int, ctx *emitterContext, term *syntax.Term) {

	if term.TermTag == syntax.VAR && term.VarType != tokens.VT_E {

		if _, ok := ctx.fixedVars[term.Name]; ok {
			return
		}
	}

	ctx.patternCtx.entryPoint = 1
	f.PrintLabel(depth+1, "case 0:")
}

func (f *Data) matchingTerm(depth int, ctx *emitterContext, terms []*syntax.Term) {
	for _, term := range terms {

		switch term.TermTag {
		case syntax.VAR:
			f.matchingVariable(depth, ctx, &term.Value)
			break
		case syntax.STR:
			f.matchingStrLiteral(depth, ctx, string(term.Value.Str))
			break
		case syntax.COMP:
			f.matchingCompLiteral(depth, ctx, term.Value.Name)
			break
		case syntax.INT:
			f.matchingIntLiteral(depth, ctx, int(term.Value.Int.Int64()))
			break
		case syntax.EXPR:
			f.matchingExpr(depth, ctx, term.Exprs[0].Terms)
			break
		case syntax.FLOAT:
			//TO DO
			break
		}
	}
}

func (f *Data) matchingExpr(depth int, ctx *emitterContext, terms []*syntax.Term) {

	f.PrintLabel(depth, "//Check (")
	f.printOffsetCheck(depth, ctx.patternCtx.prevEntryPoint, " || memMngr.vterms[fragmentOffset].tag != V_BRACKET_TAG || "+
		"memMngr.vterms[fragmentOffset].inBracketLength == 0")

	f.PrintLabel(depth, "fragmentOffset++;")

	f.matchingTerm(depth, ctx, terms)

	f.PrintLabel(depth, "//Check )")
	f.printOffsetCheck(depth, ctx.patternCtx.prevEntryPoint, " || memMngr.vterms[fragmentOffset].tag != V_BRACKET_TAG || "+
		"memMngr.vterms[fragmentOffset].inBracketLength != 0")

	f.PrintLabel(depth, "fragmentOffset++;")
}

func (f *Data) processPatternFail(depth int, ctx *emitterContext) {

	f.PrintLabel(depth, "if (stretchingVarNumber < 0)")
	f.PrintLabel(depth, "{")

	//First pattern in current sentence
	if ctx.patternNumber == 0 || ctx.prevEntryPoint == -1 {
		f.processFailOfFirstPattern(depth+1, ctx)
	} else {
		f.processFailOfCommonPattern(depth+1, ctx.entryPoint-1)
	}

	f.PrintLabel(depth+1, "break;")
	f.PrintLabel(depth, "}")
}

func (f *Data) processFailOfFirstPattern(depth int, ctx *emitterContext) {
	if ctx.isLastSentence {
		f.PrintLabel(depth, "//First pattern of last sentence -> nothing to stretch -> fail!")
		f.PrintLabel(depth, "funcRes = (struct func_result_t){.status = FAIL_RESULT, .fieldChain = 0, .callChain = 0};")
		f.PrintLabel(depth, "*entryPoint = -1;")

	} else {
		f.PrintLabel(depth, "//First pattern of current sentence -> jump to first pattern of next sentence!")
		f.PrintLabel(depth, "stretching = 0;")
		f.PrintLabel(depth, fmt.Sprintf("*entryPoint = %d;", ctx.nextSentenceEntryPoint))
		f.initSretchVarNumbers(depth, ctx.maxPatternNumber)
	}
}

func (f *Data) processFailOfCommonPattern(depth, prevEntryPoint int) {
	f.PrintLabel(depth, "//Jump to previouse pattern of same sentence!")
	f.PrintLabel(depth, fmt.Sprintf("*entryPoint = %d;", prevEntryPoint))
}

func (f *Data) initSretchVarNumbers(depth, maxPatternNumber int) {

	f.PrintLabel(depth, fmt.Sprintf("for (i = 0; i < %d; ++i )", maxPatternNumber))
	f.PrintLabel(depth+1, "env->stretchVarsNumber[i] = 0;")
}

func (f *Data) checkAndAssemblyChain(depth, patternNumber int) {
	prevPatternNumber := patternNumber - 1

	f.PrintLabel(depth, "if (!stretching)")
	f.PrintLabel(depth, "{")

	if prevPatternNumber == -1 {
		f.PrintLabel(depth+1, fmt.Sprintf("if (env->_FOVs[%d] != fieldOfView)", patternNumber))
		f.printAssemblyChain(depth+1, patternNumber)
	} else {
		f.PrintLabel(depth+1, fmt.Sprintf("if (env->_FOVs[%d] == fieldOfView)", prevPatternNumber))
		f.printGetPrevAssembledFOV(depth+1, prevPatternNumber, patternNumber)
		f.PrintLabel(depth+1, "else")
		f.printAssemblyChain(depth+1, patternNumber)
	}

	f.PrintLabel(depth, "}")
	f.PrintLabel(depth, fmt.Sprintf("currFrag = env->assembledFOVs[%d]->fragment;", patternNumber))
}

func (f *Data) printAssemblyChain(depth, entryPoint int) {
	f.PrintLabel(depth, "{")
	f.PrintLabel(depth+1, fmt.Sprintf("//WARN: Correct free env->_FOVs[%d]", entryPoint))
	f.PrintLabel(depth+1, fmt.Sprintf("env->_FOVs[%d] = fieldOfView;", entryPoint))
	f.PrintLabel(depth+1, fmt.Sprintf("env->assembledFOVs[%d] = getAssembliedChain(fieldOfView);", entryPoint))
	f.PrintLabel(depth, "}")
}

func (f *Data) printGetPrevAssembledFOV(depth, prevEntryPoint, entryPoint int) {
	f.PrintLabel(depth, "{")
	f.PrintLabel(depth+1, fmt.Sprintf("//WARN: Correct free env->_FOVs[%d]", entryPoint))
	f.PrintLabel(depth+1, fmt.Sprintf("env->_FOVs[%d] = fieldOfView;", entryPoint))
	f.PrintLabel(depth+1, fmt.Sprintf("env->assembledFOVs[%d] = env->assembledFOVs[%d];", entryPoint, prevEntryPoint))
	f.PrintLabel(depth, "}")
}

func (f *Data) matchingVariable(depth int, ctx *emitterContext, value *tokens.Value) {

	varInfo, isLocalVar := ctx.sentenceScope.VarMap[value.Name]
	isFixedVar := true
	matchedEntryPoint := 0

	if !isLocalVar {
		//Env var
		varInfo = ctx.envVarMap[value.Name]

	} else {
		matchedEntryPoint, isFixedVar = ctx.fixedVars[value.Name]
	}

	varNumber := varInfo.Number

	f.PrintLabel(depth-1, fmt.Sprintf("//Matching %s variable", value.Name))

	switch value.VarType {
	case tokens.VT_T:
		if isFixedVar {
			if isLocalVar {
				f.matchingFixedLocalExprVar(depth, ctx.patternCtx.prevEntryPoint, matchedEntryPoint, varNumber)
			} else {
				f.matchingFixedEnvExprVar(depth, ctx.patternCtx.prevEntryPoint, varNumber)
			}
		} else {
			f.matchingFreeTermVar(depth, ctx.patternCtx.prevEntryPoint, ctx.patternNumber, varNumber)
			ctx.fixedVars[value.Name] = ctx.entryPoint
		}
		break

	case tokens.VT_S:
		if isFixedVar {
			if isLocalVar {
				f.matchingFixedLocalSymbolVar(depth, ctx.patternCtx.prevEntryPoint, matchedEntryPoint, varNumber)
			} else {
				f.matchingFixedEnvSymbolVar(depth, ctx.patternCtx.prevEntryPoint, varNumber)
			}

		} else {
			f.matchingFreeSymbolVar(depth, ctx.patternCtx.prevEntryPoint, ctx.patternNumber, varNumber)
			ctx.fixedVars[value.Name] = ctx.entryPoint
		}
		break

	case tokens.VT_E:

		if isFixedVar {
			if isLocalVar {
				f.matchingFixedLocalExprVar(depth, ctx.patternCtx.prevEntryPoint, matchedEntryPoint, varNumber)
			} else {
				f.matchingFixedEnvExprVar(depth, ctx.patternCtx.prevEntryPoint, varNumber)
			}
		} else {
			f.PrintLabel(depth-1, fmt.Sprintf("case %d:", ctx.patternCtx.entryPoint))

			f.matchingFreeExprVar(depth, ctx.patternCtx.prevEntryPoint, ctx.patternNumber, varNumber)

			ctx.fixedVars[value.Name] = ctx.entryPoint
			ctx.patternCtx.prevEntryPoint = ctx.patternCtx.entryPoint
			ctx.patternCtx.entryPoint++
		}
		break
	}
}

func (f *Data) printFailBlock(depth, prevStretchVarNumber int, withBreakStatement bool) {

	f.PrintLabel(depth, "{")
	f.PrintLabel(depth+1, "stretching = 1;")
	f.PrintLabel(depth+1, fmt.Sprintf("stretchingVarNumber = %d;", prevStretchVarNumber))
	if withBreakStatement {
		f.PrintLabel(depth+1, "break;")
	}
	f.PrintLabel(depth, "}")
}

func (f *Data) printOffsetCheck(depth, prevStretchVarNumber int, optionalCond string) {
	f.PrintLabel(depth, fmt.Sprintf("if (fragmentOffset >= currFrag->offset + currFrag->length%s)", optionalCond))
	f.printFailBlock(depth, prevStretchVarNumber, true)
}
