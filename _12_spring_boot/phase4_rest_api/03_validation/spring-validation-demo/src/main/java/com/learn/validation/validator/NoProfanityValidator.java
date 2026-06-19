package com.learn.validation.validator;

import jakarta.validation.ConstraintValidator;
import jakarta.validation.ConstraintValidatorContext;

import java.util.Set;

/**
 * Custom validator implementation.
 * 
 * Implements ConstraintValidator<AnnotationType, FieldType>
 * - isValid() returns true if valid, false if invalid
 */
public class NoProfanityValidator implements ConstraintValidator<NoProfanity, String> {

    // Simulated banned words list
    private static final Set<String> BANNED_WORDS = Set.of(
        "spam", "scam", "fake", "illegal", "banned"
    );

    @Override
    public boolean isValid(String value, ConstraintValidatorContext context) {
        // null values should be handled by @NotNull / @NotBlank
        if (value == null) {
            return true;
        }

        String lower = value.toLowerCase();
        return BANNED_WORDS.stream().noneMatch(lower::contains);
    }
}
