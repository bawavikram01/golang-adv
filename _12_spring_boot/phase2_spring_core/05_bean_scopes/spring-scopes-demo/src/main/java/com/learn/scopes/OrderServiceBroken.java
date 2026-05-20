package com.learn.scopes;

import org.springframework.stereotype.Component;

/**
 * THE PROTOTYPE TRAP:
 * This singleton injects a prototype bean. But the injection happens
 * only ONCE (when the singleton is created). So the same prototype
 * instance is reused every time!
 *
 * Problem: The "cart" should be fresh for every order, but it's shared.
 */
@Component
public class OrderServiceBroken {

    private final PrototypeService cart; // Injected ONCE!

    public OrderServiceBroken(PrototypeService cart) {
        this.cart = cart;
        System.out.println("  [Broken] OrderServiceBroken created - cart injected once: " + cart);
    }

    public void placeOrder(String orderId) {
        cart.addItem(orderId);
        System.out.println("    " + orderId + " → cart items: " + cart.getItems()
                + "  (cart=" + Integer.toHexString(cart.hashCode()) + ")");
    }
}
