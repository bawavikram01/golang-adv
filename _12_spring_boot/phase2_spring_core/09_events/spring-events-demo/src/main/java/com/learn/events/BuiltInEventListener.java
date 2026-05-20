package com.learn.events;

import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.ContextClosedEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

/**
 * Listens to BUILT-IN Spring events.
 * Spring publishes these automatically during the application lifecycle.
 */
@Component
public class BuiltInEventListener {

    @EventListener
    public void onAppReady(ApplicationReadyEvent event) {
        System.out.println("  🚀 [Built-in] ApplicationReadyEvent — app fully started in "
                + event.getTimeTaken().toMillis() + "ms");
    }

    @EventListener
    public void onShutdown(ContextClosedEvent event) {
        System.out.println("  🛑 [Built-in] ContextClosedEvent — application shutting down");
    }
}
