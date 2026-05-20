package com.learn.aop;

import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

/**
 * ASPECT 2: Timing — demonstrates @Around (the most powerful advice).
 *
 * Only applies to methods annotated with @Timed (custom annotation).
 * This is the cleanest pattern: annotation-driven AOP.
 */
@Aspect
@Component
@Order(2)  // Runs AFTER LoggingAspect
public class TimingAspect {

    @Around("@annotation(com.learn.aop.Timed)")
    public Object measureExecutionTime(ProceedingJoinPoint joinPoint) throws Throwable {
        String method = joinPoint.getSignature().getName();

        long startTime = System.nanoTime();

        // ─── Execute the real method ───
        Object result = joinPoint.proceed();

        long duration = (System.nanoTime() - startTime) / 1_000_000;
        System.out.println("      [TIMING @Around] " + method + "() took " + duration + "ms");

        return result;  // Must return the result!
    }
}
