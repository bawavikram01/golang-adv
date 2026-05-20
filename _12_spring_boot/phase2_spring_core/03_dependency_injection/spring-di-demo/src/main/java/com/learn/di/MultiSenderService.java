package com.learn.di;

import org.springframework.stereotype.Service;
import java.util.List;

/**
 * INJECT ALL IMPLEMENTATIONS at once.
 *
 * When you need ALL beans of a type (not just one), inject a List<T>.
 * Spring collects ALL matching beans and gives you the full list.
 *
 * Use case: send notification on ALL channels, try multiple strategies, etc.
 */
@Service
public class MultiSenderService {

    // Spring injects ALL beans that implement MessageSender!
    private final List<MessageSender> allSenders;

    public MultiSenderService(List<MessageSender> allSenders) {
        this.allSenders = allSenders;
        System.out.println("  [MULTI-DI] Created with " + allSenders.size() + " senders: "
            + allSenders.stream().map(MessageSender::getChannel).toList());
    }

    public void broadcastToAll(String contact, String message) {
        System.out.println("\n    [MultiSenderService] Broadcasting to ALL channels:");
        for (MessageSender sender : allSenders) {
            sender.send(contact, message);
        }
    }
}
