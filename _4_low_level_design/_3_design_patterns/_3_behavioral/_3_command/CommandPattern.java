/*
 * =============================================================
 * BEHAVIORAL PATTERN 3: COMMAND
 * =============================================================
 *
 * INTENT: Encapsulate a request as an object, allowing you to
 *         parameterize, queue, log, and UNDO operations.
 *
 * ANALOGY: Restaurant — you give your ORDER (command) to the waiter.
 *          The waiter doesn't cook; they pass the command to the kitchen.
 *
 * USE WHEN:
 *   - You need undo/redo functionality
 *   - You want to queue or schedule operations
 *   - You want to log all operations
 *   - You want to decouple "what" from "who does it"
 *
 * REAL EXAMPLES: javax.swing.Action, Runnable, undo in editors
 */

import java.util.*;

public class CommandPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Text Editor with Undo/Redo
        // ═══════════════════════════════════════════════════════
        System.out.println("=== TEXT EDITOR WITH UNDO/REDO ===");

        TextEditor editor = new TextEditor();
        CommandManager manager = new CommandManager();

        // Execute commands
        manager.execute(new TypeCommand(editor, "Hello"));
        manager.execute(new TypeCommand(editor, " World"));
        manager.execute(new TypeCommand(editor, "!"));
        System.out.println("  Text: \"" + editor.getText() + "\"");

        // Undo last two
        manager.undo();
        System.out.println("  After undo: \"" + editor.getText() + "\"");
        manager.undo();
        System.out.println("  After undo: \"" + editor.getText() + "\"");

        // Redo one
        manager.redo();
        System.out.println("  After redo: \"" + editor.getText() + "\"");

        // ═══════════════════════════════════════════════════════
        // Smart Home Remote Control
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== SMART HOME REMOTE ===");

        Light livingRoomLight = new Light("Living Room");
        Fan ceilingFan = new Fan("Ceiling");
        Thermostat thermostat = new Thermostat();

        RemoteControl remote = new RemoteControl();

        // Program the remote — any button can do any command!
        remote.setCommand(0, new LightOnCommand(livingRoomLight), new LightOffCommand(livingRoomLight));
        remote.setCommand(1, new FanHighCommand(ceilingFan), new FanOffCommand(ceilingFan));
        remote.setCommand(2, new ThermostatSetCommand(thermostat, 72), new ThermostatSetCommand(thermostat, 68));

        // Press buttons
        remote.onButtonPressed(0);   // Light on
        remote.onButtonPressed(1);   // Fan high
        remote.onButtonPressed(2);   // Thermostat 72

        System.out.println();
        remote.undoButtonPressed();   // Undo last (thermostat back)

        System.out.println();
        remote.offButtonPressed(0);  // Light off
        remote.offButtonPressed(1);  // Fan off

        // ═══════════════════════════════════════════════════════
        // Macro Command — composite command
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== MACRO COMMAND ===");
        Command partyMode = new MacroCommand(
                new LightOnCommand(livingRoomLight),
                new FanHighCommand(ceilingFan),
                new ThermostatSetCommand(thermostat, 75)
        );
        System.out.println("Activating party mode...");
        partyMode.execute();
    }
}

// ═══════════════════════════════════════════════════════════════
// COMMAND INTERFACE
// ═══════════════════════════════════════════════════════════════
interface Command {
    void execute();
    void undo();
}

// ═══════════════════════════════════════════════════════════════
// TEXT EDITOR (Receiver)
// ═══════════════════════════════════════════════════════════════
class TextEditor {
    private StringBuilder text = new StringBuilder();

    public void type(String str) {
        text.append(str);
    }

    public void delete(int length) {
        text.delete(text.length() - length, text.length());
    }

    public String getText() {
        return text.toString();
    }
}

class TypeCommand implements Command {
    private TextEditor editor;
    private String textToType;

    public TypeCommand(TextEditor editor, String text) {
        this.editor = editor;
        this.textToType = text;
    }

    @Override
    public void execute() {
        editor.type(textToType);
        System.out.println("  Typed: \"" + textToType + "\"");
    }

    @Override
    public void undo() {
        editor.delete(textToType.length());
        System.out.println("  Undo typing: \"" + textToType + "\"");
    }
}

// Command Manager — handles undo/redo stacks
class CommandManager {
    private Deque<Command> undoStack = new ArrayDeque<>();
    private Deque<Command> redoStack = new ArrayDeque<>();

    public void execute(Command cmd) {
        cmd.execute();
        undoStack.push(cmd);
        redoStack.clear();  // new action invalidates redo history
    }

    public void undo() {
        if (undoStack.isEmpty()) {
            System.out.println("  Nothing to undo!");
            return;
        }
        Command cmd = undoStack.pop();
        cmd.undo();
        redoStack.push(cmd);
    }

    public void redo() {
        if (redoStack.isEmpty()) {
            System.out.println("  Nothing to redo!");
            return;
        }
        Command cmd = redoStack.pop();
        cmd.execute();
        undoStack.push(cmd);
    }
}

// ═══════════════════════════════════════════════════════════════
// SMART HOME (Receivers)
// ═══════════════════════════════════════════════════════════════
class Light {
    private String location;
    private boolean on = false;

    public Light(String location) { this.location = location; }

    public void turnOn()  { on = true;  System.out.println("  💡 " + location + " light ON"); }
    public void turnOff() { on = false; System.out.println("  💡 " + location + " light OFF"); }
}

class Fan {
    private String location;
    private int speed = 0;

    public Fan(String location) { this.location = location; }

    public void setSpeed(int speed) {
        int prev = this.speed;
        this.speed = speed;
        System.out.println("  🌀 " + location + " fan: " + prev + " → " + speed);
    }

    public int getSpeed() { return speed; }
}

class Thermostat {
    private int temperature = 68;

    public void setTemperature(int temp) {
        int prev = this.temperature;
        this.temperature = temp;
        System.out.println("  🌡️ Thermostat: " + prev + "°F → " + temp + "°F");
    }

    public int getTemperature() { return temperature; }
}

// ═══════════════════════════════════════════════════════════════
// CONCRETE COMMANDS
// ═══════════════════════════════════════════════════════════════
class LightOnCommand implements Command {
    private Light light;
    public LightOnCommand(Light light) { this.light = light; }
    @Override public void execute() { light.turnOn(); }
    @Override public void undo()    { light.turnOff(); }
}

class LightOffCommand implements Command {
    private Light light;
    public LightOffCommand(Light light) { this.light = light; }
    @Override public void execute() { light.turnOff(); }
    @Override public void undo()    { light.turnOn(); }
}

class FanHighCommand implements Command {
    private Fan fan;
    private int previousSpeed;
    public FanHighCommand(Fan fan) { this.fan = fan; }
    @Override public void execute() { previousSpeed = fan.getSpeed(); fan.setSpeed(3); }
    @Override public void undo()    { fan.setSpeed(previousSpeed); }
}

class FanOffCommand implements Command {
    private Fan fan;
    private int previousSpeed;
    public FanOffCommand(Fan fan) { this.fan = fan; }
    @Override public void execute() { previousSpeed = fan.getSpeed(); fan.setSpeed(0); }
    @Override public void undo()    { fan.setSpeed(previousSpeed); }
}

class ThermostatSetCommand implements Command {
    private Thermostat thermostat;
    private int newTemp;
    private int previousTemp;
    public ThermostatSetCommand(Thermostat thermostat, int temp) { this.thermostat = thermostat; this.newTemp = temp; }
    @Override public void execute() { previousTemp = thermostat.getTemperature(); thermostat.setTemperature(newTemp); }
    @Override public void undo()    { thermostat.setTemperature(previousTemp); }
}

// ═══════════════════════════════════════════════════════════════
// MACRO COMMAND — execute multiple commands at once
// ═══════════════════════════════════════════════════════════════
class MacroCommand implements Command {
    private Command[] commands;

    public MacroCommand(Command... commands) {
        this.commands = commands;
    }

    @Override
    public void execute() {
        for (Command c : commands) c.execute();
    }

    @Override
    public void undo() {
        // Undo in REVERSE order
        for (int i = commands.length - 1; i >= 0; i--) {
            commands[i].undo();
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// REMOTE CONTROL (Invoker)
// ═══════════════════════════════════════════════════════════════
class RemoteControl {
    private Command[] onCommands = new Command[5];
    private Command[] offCommands = new Command[5];
    private Command lastCommand;

    public void setCommand(int slot, Command onCmd, Command offCmd) {
        onCommands[slot] = onCmd;
        offCommands[slot] = offCmd;
    }

    public void onButtonPressed(int slot) {
        if (onCommands[slot] != null) {
            onCommands[slot].execute();
            lastCommand = onCommands[slot];
        }
    }

    public void offButtonPressed(int slot) {
        if (offCommands[slot] != null) {
            offCommands[slot].execute();
            lastCommand = offCommands[slot];
        }
    }

    public void undoButtonPressed() {
        if (lastCommand != null) {
            System.out.println("  [UNDO]");
            lastCommand.undo();
        }
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ 4 participants: Command, ConcreteCommand, Receiver, Invoker.
 * ✦ Command encapsulates action + undo as an OBJECT.
 * ✦ Enables: undo/redo, queuing, logging, transactions.
 * ✦ MacroCommand = composite of commands.
 * ✦ Invoker doesn't know what the command does — just calls execute().
 *
 * COMPILE & RUN:
 *   javac CommandPattern.java && java CommandPattern
 */
