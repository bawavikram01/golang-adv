package com.learn.aop;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Custom annotation — methods annotated with @Timed will be
 * automatically measured by our TimingAspect.
 *
 * This is the cleanest AOP pattern:
 *   1. Define an annotation
 *   2. Apply it to methods
 *   3. Write an aspect that targets it
 */
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface Timed {
}
