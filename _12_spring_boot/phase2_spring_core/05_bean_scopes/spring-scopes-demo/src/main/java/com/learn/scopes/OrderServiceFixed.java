package com.learn.scopes;

import org.springframework.beans.factory.ObjectFactory;
import org.springframework.stereotype.Component;

/**
 * FIX for the Prototype Trap: Use ObjectFactory<T>.
 *
 * Instead of injecting the prototype directly, inject an ObjectFactory.
 * Each call to factory.getObject() asks the container for a FRESH prototype.
 *
 * Alternatives that also work:
 * - ObjectProvider<T> (Spring-specific, more features)
 * - Provider<T> (jakarta.inject standard)
 * - @Lookup annotation on a method
 * - ApplicationContext.getBean() (tight coupling, avoid)
 */
@Component
public class OrderServiceFixed {

    private final ObjectFactory<PrototypeService> cartFactory;

    public OrderServiceFixed(ObjectFactory<PrototypeService> cartFactory) {
        this.cartFactory = cartFactory;
    }

    public void placeOrder(String orderId) {
        // Each call gets a FRESH prototype instance!
        PrototypeService cart = cartFactory.getObject();
        cart.addItem(orderId);
        System.out.println("    " + orderId + " → cart items: " + cart.getItems()
                + "  (cart=" + Integer.toHexString(cart.hashCode()) + ")");
    }
}
