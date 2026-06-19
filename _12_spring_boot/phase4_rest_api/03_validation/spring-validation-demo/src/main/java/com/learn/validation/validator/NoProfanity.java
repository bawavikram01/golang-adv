package com.learn.validation.validator;

import jakarta.validation.Constraint;
import jakarta.validation.Payload;

import java.lang.annotation.*;

/**
 * Custom validation annotation.
 * Usage: @NoProfanity on a String field.
 */
@Target({ElementType.FIELD, ElementType.PARAMETER})
@Retention(RetentionPolicy.RUNTIME)
@Documented
@Constraint(validatedBy = NoProfanityValidator.class)
public @interface NoProfanity {
    String message() default "Contains inappropriate or banned words";
    Class<?>[] groups() default {};
    Class<? extends Payload>[] payload() default {};
}
