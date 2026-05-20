package com.learn.aop;

import org.aspectj.lang.JoinPoint;
import org.aspectj.lang.annotation.*;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

import java.util.Arrays;

/**
 * ASPECT 1: Logging — demonstrates @Before, @AfterReturning, @AfterThrowing, @After.
 *
 * Applies to ALL public methods in the com.learn.aop package (except aspects).
 * This is the classic "method-level logging" use case.
 */
@Aspect
@Component
@Order(1)  // Runs FIRST among aspects
public class LoggingAspect {

    // ─── Named Pointcut (reusable) ──────────────────────────────
    @Pointcut("execution(public * com.learn.aop.*Service.*(..))")
    public void allServiceMethods() {}

    // ─── @Before — runs before the method ───────────────────────
    @Before("allServiceMethods()")
    public void logBefore(JoinPoint jp) {
        String className = jp.getTarget().getClass().getSimpleName();
        String method = jp.getSignature().getName();
        Object[] args = jp.getArgs();

        System.out.println("      [LOG @Before] " + className + "." + method
                + "(" + formatArgs(args) + ")");
    }

    // ─── @AfterReturning — runs after successful return ─────────
    @AfterReturning(pointcut = "allServiceMethods()", returning = "result")
    public void logAfterReturning(JoinPoint jp, Object result) {
        String method = jp.getSignature().getName();
        System.out.println("      [LOG @AfterReturning] " + method + " → returned: " + result);
    }

    // ─── @AfterThrowing — runs only if exception thrown ─────────
    @AfterThrowing(pointcut = "allServiceMethods()", throwing = "ex")
    public void logAfterThrowing(JoinPoint jp, Throwable ex) {
        String method = jp.getSignature().getName();
        System.out.println("      [LOG @AfterThrowing] " + method + " → EXCEPTION: " + ex.getMessage());
    }

    // ─── @After — runs always (like finally) ────────────────────
    @After("allServiceMethods()")
    public void logAfter(JoinPoint jp) {
        String method = jp.getSignature().getName();
        System.out.println("      [LOG @After] " + method + " completed (finally)");
    }

    private String formatArgs(Object[] args) {
        if (args.length == 0) return "";
        return Arrays.stream(args)
                .map(a -> a instanceof String ? "\"" + a + "\"" : String.valueOf(a))
                .reduce((a, b) -> a + ", " + b)
                .orElse("");
    }
}
