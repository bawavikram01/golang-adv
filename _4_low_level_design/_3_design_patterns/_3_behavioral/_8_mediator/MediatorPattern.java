/*
 * =============================================================
 * BEHAVIORAL PATTERN 8: MEDIATOR
 * =============================================================
 *
 * INTENT: Define an object that encapsulates how a set of objects
 *         interact. Instead of objects talking directly to each other,
 *         they communicate through a MEDIATOR.
 *
 * ANALOGY: Airport Control Tower — planes don't talk to each other
 *          directly. They all communicate through the control tower.
 *          The tower coordinates takeoffs, landings, and collisions.
 *
 * USE WHEN:
 *   - Many objects communicate in complex ways (spaghetti dependencies)
 *   - You want to reduce coupling between communicating objects
 *   - You want to centralize control logic
 *
 * REAL EXAMPLES: Chat rooms, event buses, MVC (Controller),
 *                Dialog boxes (button clicks update text fields)
 *
 * MEDIATOR vs OBSERVER:
 *   - Observer: one-to-many notification (unidirectional)
 *   - Mediator: many-to-many coordination (bidirectional)
 */

import java.util.*;

public class MediatorPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Chat Room Mediator
        // ═══════════════════════════════════════════════════════
        System.out.println("=== CHAT ROOM MEDIATOR ===");

        ChatRoom room = new ChatRoom("Java Devs");

        ChatUser alice = new ChatUser("Alice", room);
        ChatUser bob = new ChatUser("Bob", room);
        ChatUser charlie = new ChatUser("Charlie", room);

        alice.send("Hey everyone!");
        bob.send("Hi Alice!");
        charlie.send("What's up?");

        System.out.println();
        bob.sendPrivate("alice", "Can you review my PR?");

        // ═══════════════════════════════════════════════════════
        // Airport Control Tower
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== AIR TRAFFIC CONTROL MEDIATOR ===");

        ControlTower tower = new ControlTower();

        Aircraft flight1 = new Aircraft("FL-101", tower);
        Aircraft flight2 = new Aircraft("FL-202", tower);
        Aircraft flight3 = new Aircraft("FL-303", tower);

        flight1.requestLanding();
        flight2.requestLanding();  // runway busy — queued
        flight1.completeLanding();  // runway freed — next plane cleared
        flight3.requestTakeoff();

        // ═══════════════════════════════════════════════════════
        // Smart Home Mediator
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== SMART HOME MEDIATOR ===");

        SmartHomeMediator home = new SmartHomeMediator();

        Sensor motionSensor = new Sensor("Motion Sensor", home);
        Sensor tempSensor = new Sensor("Temp Sensor", home);
        SmartLight light = new SmartLight("Living Room Light", home);
        SmartAC ac = new SmartAC("Central AC", home);
        Alarm alarm = new Alarm("Home Alarm", home);

        home.registerDevice("motion", motionSensor);
        home.registerDevice("temp", tempSensor);
        home.registerDevice("light", light);
        home.registerDevice("ac", ac);
        home.registerDevice("alarm", alarm);

        System.out.println("--- Motion detected at night ---");
        motionSensor.trigger("MOTION_DETECTED_NIGHT");

        System.out.println("\n--- Temperature too high ---");
        tempSensor.trigger("TEMP_HIGH");

        System.out.println("\n--- Motion detected while away ---");
        motionSensor.trigger("MOTION_DETECTED_AWAY");
    }
}

// ═══════════════════════════════════════════════════════════════
// CHAT ROOM MEDIATOR
// ═══════════════════════════════════════════════════════════════
interface ChatMediator {
    void sendMessage(String message, ChatUser sender);
    void sendPrivateMessage(String message, ChatUser sender, String recipientName);
    void addUser(ChatUser user);
}

class ChatRoom implements ChatMediator {
    private String name;
    private Map<String, ChatUser> users = new HashMap<>();

    public ChatRoom(String name) { this.name = name; }

    @Override
    public void addUser(ChatUser user) {
        users.put(user.getName().toLowerCase(), user);
        System.out.println("  [" + name + "] " + user.getName() + " joined.");
    }

    @Override
    public void sendMessage(String message, ChatUser sender) {
        for (ChatUser user : users.values()) {
            if (user != sender) {  // don't send to self
                user.receive(message, sender.getName());
            }
        }
    }

    @Override
    public void sendPrivateMessage(String message, ChatUser sender, String recipientName) {
        ChatUser recipient = users.get(recipientName.toLowerCase());
        if (recipient != null) {
            recipient.receive("[DM] " + message, sender.getName());
        } else {
            System.out.println("  User '" + recipientName + "' not found.");
        }
    }
}

class ChatUser {
    private String name;
    private ChatMediator mediator;

    public ChatUser(String name, ChatMediator mediator) {
        this.name = name;
        this.mediator = mediator;
        mediator.addUser(this);
    }

    public void send(String message) {
        System.out.println("  " + name + " says: " + message);
        mediator.sendMessage(message, this);
    }

    public void sendPrivate(String to, String message) {
        System.out.println("  " + name + " → " + to + " (private): " + message);
        mediator.sendPrivateMessage(message, this, to);
    }

    public void receive(String message, String from) {
        System.out.println("    📩 " + name + " received from " + from + ": " + message);
    }

    public String getName() { return name; }
}

// ═══════════════════════════════════════════════════════════════
// AIRPORT CONTROL TOWER
// ═══════════════════════════════════════════════════════════════
class ControlTower {
    private boolean runwayFree = true;
    private Queue<Aircraft> landingQueue = new LinkedList<>();

    public void requestLanding(Aircraft aircraft) {
        if (runwayFree) {
            runwayFree = false;
            System.out.println("  🛬 " + aircraft.getId() + ": Cleared to land.");
        } else {
            landingQueue.add(aircraft);
            System.out.println("  ⏳ " + aircraft.getId() + ": Runway busy. Queued (pos " + landingQueue.size() + ")");
        }
    }

    public void requestTakeoff(Aircraft aircraft) {
        if (runwayFree) {
            System.out.println("  🛫 " + aircraft.getId() + ": Cleared for takeoff.");
        } else {
            System.out.println("  ⏳ " + aircraft.getId() + ": Runway busy. Wait.");
        }
    }

    public void notifyLandingComplete(Aircraft aircraft) {
        System.out.println("  ✓ " + aircraft.getId() + ": Landing complete. Runway freed.");
        runwayFree = true;

        if (!landingQueue.isEmpty()) {
            Aircraft next = landingQueue.poll();
            requestLanding(next);
        }
    }
}

class Aircraft {
    private String id;
    private ControlTower tower;

    public Aircraft(String id, ControlTower tower) {
        this.id = id;
        this.tower = tower;
    }

    public void requestLanding()   { tower.requestLanding(this); }
    public void requestTakeoff()   { tower.requestTakeoff(this); }
    public void completeLanding()  { tower.notifyLandingComplete(this); }
    public String getId()          { return id; }
}

// ═══════════════════════════════════════════════════════════════
// SMART HOME MEDIATOR
// ═══════════════════════════════════════════════════════════════
interface SmartDevice {
    String getId();
    void onEvent(String event);
}

class SmartHomeMediator {
    private Map<String, SmartDevice> devices = new HashMap<>();

    public void registerDevice(String key, SmartDevice device) {
        devices.put(key, device);
    }

    public void notify(String event, SmartDevice source) {
        // Central coordination logic — devices don't know about each other
        switch (event) {
            case "MOTION_DETECTED_NIGHT":
                getDevice("light").onEvent("TURN_ON");
                break;
            case "MOTION_DETECTED_AWAY":
                getDevice("alarm").onEvent("TRIGGER");
                getDevice("light").onEvent("FLASH");
                break;
            case "TEMP_HIGH":
                getDevice("ac").onEvent("COOL_DOWN");
                break;
        }
    }

    private SmartDevice getDevice(String key) {
        return devices.get(key);
    }
}

class Sensor implements SmartDevice {
    private String name;
    private SmartHomeMediator mediator;

    public Sensor(String name, SmartHomeMediator mediator) {
        this.name = name;
        this.mediator = mediator;
    }

    public void trigger(String event) {
        System.out.println("  📡 " + name + " triggered: " + event);
        mediator.notify(event, this);
    }

    @Override public String getId() { return name; }
    @Override public void onEvent(String event) {}
}

class SmartLight implements SmartDevice {
    private String name;
    private SmartHomeMediator mediator;

    public SmartLight(String name, SmartHomeMediator mediator) {
        this.name = name;
        this.mediator = mediator;
    }

    @Override public String getId() { return name; }

    @Override
    public void onEvent(String event) {
        switch (event) {
            case "TURN_ON" -> System.out.println("  💡 " + name + ": Turning ON");
            case "TURN_OFF" -> System.out.println("  💡 " + name + ": Turning OFF");
            case "FLASH" -> System.out.println("  🚨 " + name + ": FLASHING (intruder alert)");
        }
    }
}

class SmartAC implements SmartDevice {
    private String name;
    private SmartHomeMediator mediator;

    public SmartAC(String name, SmartHomeMediator mediator) {
        this.name = name;
        this.mediator = mediator;
    }

    @Override public String getId() { return name; }

    @Override
    public void onEvent(String event) {
        if ("COOL_DOWN".equals(event)) {
            System.out.println("  ❄️ " + name + ": Cooling down to 22°C");
        }
    }
}

class Alarm implements SmartDevice {
    private String name;
    private SmartHomeMediator mediator;

    public Alarm(String name, SmartHomeMediator mediator) {
        this.name = name;
        this.mediator = mediator;
    }

    @Override public String getId() { return name; }

    @Override
    public void onEvent(String event) {
        if ("TRIGGER".equals(event)) {
            System.out.println("  🚨 " + name + ": ALARM TRIGGERED!");
        }
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Mediator centralizes communication between objects.
 * ✦ Objects don't reference each other — only the mediator.
 * ✦ Reduces N×N dependencies to N×1 (star topology).
 *
 * ✦ Common uses:
 *   - Chat rooms (users communicate through room)
 *   - Air traffic control (planes through tower)
 *   - GUI dialog boxes (widgets through dialog)
 *   - Smart home (devices through hub)
 *
 * ✦ Trade-off: mediator can become a "God object" if too much
 *   logic is centralized. Keep it focused.
 *
 * COMPILE & RUN:
 *   javac MediatorPattern.java && java MediatorPattern
 */
