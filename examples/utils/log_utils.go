package utils

import (
	"log"

	"github.com/seyedali-dev/gosaidsno/aspect"
)

func LogBefore(ctx *aspect.Context, priority int, message string) {
	log.Printf("🟢 [BEFORE] %s - Priority: %d [%s]", ctx.FunctionName, priority, message)
}

func LogAfter(ctx *aspect.Context, priority int, message string) {
	log.Printf("🔵 [AFTER] %s - Priority: %d [%s]", ctx.FunctionName, priority, message)
}

func LogAround(ctx *aspect.Context, priority int, message string) {
	log.Printf("🟠 [AROUND] %s - Priority: %d [%s]", ctx.FunctionName, priority, message)
}

func LogAfterReturning(ctx *aspect.Context, priority int, message string) {
	log.Printf("🟣 [AFTER_RETURNING] %s - Priority: %d [%s]", ctx.FunctionName, priority, message)
}

func LogAfterThrowing(ctx *aspect.Context, priority int, message string) {
	log.Printf("🔴 [AFTER_THROWING] %s - Priority: %d [%s]", ctx.FunctionName, priority, message)
}
