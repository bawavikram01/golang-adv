package com.learn.exceptions.exception;

/**
 * Thrown when a business rule is violated (e.g., insufficient balance).
 * Maps to HTTP 422 Unprocessable Entity.
 */
public class BusinessRuleException extends BusinessException {

    public BusinessRuleException(String message) {
        super(message, "BUSINESS_RULE_VIOLATION");
    }
}
