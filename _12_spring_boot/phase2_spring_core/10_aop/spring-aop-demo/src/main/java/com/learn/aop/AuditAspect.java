package com.learn.aop;

import org.aspectj.lang.JoinPoint;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Before;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

/**
 * ASPECT 3: Audit — demonstrates targeting a custom annotation with parameters.
 *
 * Applies only to methods annotated with @Audited.
 * Reads the annotation's 'action' parameter.
 */
@Aspect
@Component
@Order(3)
public class AuditAspect {

    @Before("@annotation(audited)")
    public void auditAction(JoinPoint jp, Audited audited) {
        String className = jp.getTarget().getClass().getSimpleName();
        String method = jp.getSignature().getName();
        String action = audited.action().isEmpty() ? method : audited.action();

        System.out.println("      [AUDIT] Action=" + action + " | " + className + "." + method + "()");
    }
}
