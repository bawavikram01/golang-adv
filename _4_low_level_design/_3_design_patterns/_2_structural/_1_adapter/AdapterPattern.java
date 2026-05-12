/*
 * =============================================================
 * STRUCTURAL PATTERN 1: ADAPTER
 * =============================================================
 *
 * INTENT: Convert the interface of a class into another interface
 *         that clients expect. Makes incompatible interfaces work together.
 *
 * ANALOGY: Travel power adapter — your US plug (interface A) works
 *          in a European socket (interface B) through an adapter.
 *
 * USE WHEN:
 *   - Integrating a third-party library with incompatible interface
 *   - Wrapping legacy code to work with new systems
 *   - Unifying different APIs under one interface
 */

import java.util.List;

public class AdapterPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Scenario: Our app uses MediaPlayer interface.
        // We want to play MP4 and VLC files using third-party libraries.
        // The third-party libraries have DIFFERENT interfaces!
        // ═══════════════════════════════════════════════════════
        System.out.println("=== ADAPTER PATTERN ===");

        // Our standard player — plays MP3 natively
        MediaPlayer mp3Player = new Mp3Player();
        mp3Player.play("song.mp3");

        // Third-party players have incompatible interfaces
        // AdvancedVideoPlayer.playMp4() — different method name/signature
        // VlcVideoPlayer.playVlc() — different method name/signature

        // ADAPTERS make them compatible with our MediaPlayer interface
        MediaPlayer mp4Player = new Mp4PlayerAdapter(new AdvancedVideoPlayer());
        MediaPlayer vlcPlayer = new VlcPlayerAdapter(new VlcVideoPlayer());

        mp4Player.play("movie.mp4");
        vlcPlayer.play("lecture.vlc");

        // Now we can treat ALL players uniformly!
        System.out.println("\n=== UNIFORM PLAYER LIST ===");
        List<MediaPlayer> players = List.of(mp3Player, mp4Player, vlcPlayer);
        String[] files = {"track.mp3", "video.mp4", "clip.vlc"};
        for (int i = 0; i < players.size(); i++) {
            players.get(i).play(files[i]);
        }

        // ═══════════════════════════════════════════════════════
        // Real-world: Payment Gateway Adapter
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== REAL-WORLD: Payment Adapter ===");

        PaymentProcessor stripe = new StripeAdapter(new StripeAPI());
        PaymentProcessor paypal = new PayPalAdapter(new PayPalAPI());

        stripe.pay(99.99);
        paypal.pay(49.99);
    }
}

// ═══════════════════════════════════════════════════════════════
// TARGET: The interface our code expects
// ═══════════════════════════════════════════════════════════════
interface MediaPlayer {
    void play(String filename);
}

// Our native implementation
class Mp3Player implements MediaPlayer {
    @Override
    public void play(String filename) {
        System.out.println("  🎵 Playing MP3: " + filename);
    }
}

// ═══════════════════════════════════════════════════════════════
// ADAPTEES: Third-party classes with INCOMPATIBLE interfaces
// ═══════════════════════════════════════════════════════════════
class AdvancedVideoPlayer {
    // Different method name and no interface!
    public void playMp4File(String filename) {
        System.out.println("  🎬 Advanced Player: Playing MP4 → " + filename);
    }
}

class VlcVideoPlayer {
    // Completely different API
    public void loadMedia(String path) {
        System.out.println("  📀 VLC: Loading → " + path);
    }
    public void startPlayback() {
        System.out.println("  📀 VLC: Playback started");
    }
}

// ═══════════════════════════════════════════════════════════════
// ADAPTERS: Bridge between our interface and third-party classes
// ═══════════════════════════════════════════════════════════════
class Mp4PlayerAdapter implements MediaPlayer {
    private AdvancedVideoPlayer adaptee;

    public Mp4PlayerAdapter(AdvancedVideoPlayer player) {
        this.adaptee = player;
    }

    @Override
    public void play(String filename) {
        // Translate our interface call to the adaptee's method
        adaptee.playMp4File(filename);
    }
}

class VlcPlayerAdapter implements MediaPlayer {
    private VlcVideoPlayer adaptee;

    public VlcPlayerAdapter(VlcVideoPlayer player) {
        this.adaptee = player;
    }

    @Override
    public void play(String filename) {
        // Adapt: our single play() → their multi-step process
        adaptee.loadMedia(filename);
        adaptee.startPlayback();
    }
}

// ═══════════════════════════════════════════════════════════════
// REAL-WORLD: Payment Gateway Adapters
// ═══════════════════════════════════════════════════════════════
interface PaymentProcessor {
    void pay(double amount);
}

// Third-party: Stripe has its own API
class StripeAPI {
    public void createCharge(int amountInCents, String currency) {
        System.out.println("  💳 Stripe: Charged " + amountInCents + " " + currency);
    }
}

// Third-party: PayPal has a completely different API
class PayPalAPI {
    public void sendPayment(String amountStr) {
        System.out.println("  🅿️ PayPal: Sent $" + amountStr);
    }
}

class StripeAdapter implements PaymentProcessor {
    private StripeAPI stripe;

    public StripeAdapter(StripeAPI stripe) { this.stripe = stripe; }

    @Override
    public void pay(double amount) {
        // Convert dollars to cents, add currency
        stripe.createCharge((int)(amount * 100), "USD");
    }
}

class PayPalAdapter implements PaymentProcessor {
    private PayPalAPI paypal;

    public PayPalAdapter(PayPalAPI paypal) { this.paypal = paypal; }

    @Override
    public void pay(double amount) {
        paypal.sendPayment(String.format("%.2f", amount));
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Adapter wraps an incompatible class to match your interface.
 * ✦ Adapter HAS-A adaptee (composition, not inheritance).
 * ✦ Client code never sees the third-party API directly.
 * ✦ Perfect for integrating libraries, legacy code, or external APIs.
 * ✦ Follows OCP: add new adapters without changing existing code.
 * ✦ Follows DIP: client depends on interface, not concrete class.
 *
 * COMPILE & RUN:
 *   javac AdapterPattern.java && java AdapterPattern
 */
