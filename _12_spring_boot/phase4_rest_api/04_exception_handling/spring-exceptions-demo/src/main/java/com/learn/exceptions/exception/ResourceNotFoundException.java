package com.learn.exceptions.exception;

/**
 * Thrown when a requested resource doesn't exist.
 * Maps to HTTP 404.
 */
public class ResourceNotFoundException extends BusinessException {

    private final String resource;
    private final Object id;

    public ResourceNotFoundException(String resource, Object id) {
        super(resource + " not found with id: " + id, "RESOURCE_NOT_FOUND");
        this.resource = resource;
        this.id = id;
    }

    public String getResource() { return resource; }
    public Object getId() { return id; }
}
