package com.learn.scopes;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.context.annotation.Scope;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

/**
 * PROTOTYPE scope.
 * - A NEW instance is created every time it's requested from the container.
 * - Each instance has its own state (not shared).
 * - @PreDestroy is NOT called! Spring doesn't manage prototype destruction.
 */
@Component
@Scope(ConfigurableBeanFactory.SCOPE_PROTOTYPE)
public class PrototypeService {

    private static int instanceCount = 0;
    private final int id;
    private final List<String> items = new ArrayList<>();

    public PrototypeService() {
        this.id = ++instanceCount;
        System.out.println("  [Prototype] Constructor called → instance #" + id + " created");
    }

    public void addItem(String item) {
        items.add(item);
    }

    public List<String> getItems() {
        return items;
    }

    @Override
    public String toString() {
        return "PrototypeService@" + Integer.toHexString(hashCode()) + " (instance #" + id + ")";
    }
}
