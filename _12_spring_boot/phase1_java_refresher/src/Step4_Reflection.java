import java.lang.reflect.*;

/**
 * PHASE 1.1 — REFLECTION -  How Spring Sees Your Code
 * 
 * Theory
Reflection lets Java inspect and manipulate classes, methods, and fields at runtime — even private ones. Spring uses reflection to:

Discover which classes have @Component
Read constructor parameters to figure out dependencies
Inject values into private fields (@Autowired)
Invoke methods dynamically
Analogy
Imagine you're a building inspector. You can walk into any room (class), open any drawer (field), read any document (method), even if they're locked (private). That's reflection — Java's X-ray vision.

 * 
 * Spring uses reflection to:
 *   1. Scan classes for annotations
 *   2. Create objects (beans) without you calling "new"
 *   3. Inject dependencies into private fields
 *   4. Invoke methods dynamically
 * 
 * This is HOW Spring does its "magic". After this, no magic remains.
 */

// ============================================================
// A sample class — imagine Spring is inspecting this
// ============================================================

class PaymentService {
    private String provider = "Stripe";
    private double balance = 1000.0;

    public PaymentService() {
        // Default constructor — Spring calls this via reflection
    }

    public PaymentService(String provider) {
        this.provider = provider;
    }

    public void processPayment(double amount) {
        balance -= amount;
        System.out.println("  Paid $" + amount + " via " + provider + " | Balance: $" + balance);
    }

    private void secretMethod() {
        System.out.println("  🔒 This is private! But reflection can call me.");
    }

    public String getProvider() { return provider; }
    public double getBalance() { return balance; }
}


public class Step4_Reflection {
    public static void main(String[] args) throws Exception {

        // ============================================================
        // 1. INSPECT A CLASS — see its structure
        // ============================================================
        System.out.println("=== 1. INSPECTING PaymentService ===");
        Class<?> clazz = PaymentService.class;

        System.out.println("Class name: " + clazz.getSimpleName());

        System.out.println("\nConstructors:");
        for (Constructor<?> c : clazz.getDeclaredConstructors()) {
            System.out.println("  " + c);
        }

        System.out.println("\nFields:");
        for (Field f : clazz.getDeclaredFields()) {
            System.out.println("  " + f.getType().getSimpleName() + " " + f.getName()
                + " (private=" + Modifier.isPrivate(f.getModifiers()) + ")");
        }

        System.out.println("\nMethods:");
        for (Method m : clazz.getDeclaredMethods()) {
            System.out.println("  " + m.getName() + "(" + 
                (m.getParameterCount() > 0 ? m.getParameterTypes()[0].getSimpleName() : "") + ")");
        }


        // ============================================================
        // 2. CREATE AN OBJECT WITHOUT "new" — Spring does this!
        // ============================================================
        System.out.println("\n=== 2. CREATING OBJECT VIA REFLECTION ===");

        // Using default constructor
        Object obj1 = clazz.getDeclaredConstructor().newInstance();
        System.out.println("Created: " + obj1.getClass().getSimpleName());

        // Using parameterized constructor
        Constructor<?> paramCtor = clazz.getDeclaredConstructor(String.class);
        Object obj2 = paramCtor.newInstance("PayPal");
        System.out.println("Created with provider: " + ((PaymentService) obj2).getProvider());


        // ============================================================
        // 3. ACCESS PRIVATE FIELDS — Spring @Autowired does this!
        // ============================================================
        System.out.println("\n=== 3. ACCESSING PRIVATE FIELDS ===");

        PaymentService service = new PaymentService();
        System.out.println("Provider (via getter): " + service.getProvider()); // "Stripe"

        // Now change it via reflection — WITHOUT a setter!
        Field providerField = clazz.getDeclaredField("provider");
        providerField.setAccessible(true);  // Bypass the "private" keyword!
        providerField.set(service, "Razorpay");  // Directly set the value!

        System.out.println("Provider (after reflection): " + service.getProvider()); // "Razorpay"
        System.out.println("^ Spring's @Autowired does exactly this to inject dependencies!");


        // ============================================================
        // 4. INVOKE METHODS DYNAMICALLY
        // ============================================================
        System.out.println("\n=== 4. INVOKING METHODS VIA REFLECTION ===");

        // Call a public method
        Method processMethod = clazz.getDeclaredMethod("processPayment", double.class);
        processMethod.invoke(service, 250.0);  // Calls service.processPayment(250.0)

        // Call a PRIVATE method!
        Method secretMethod = clazz.getDeclaredMethod("secretMethod");
        secretMethod.setAccessible(true);  // Bypass private access!
        secretMethod.invoke(service);


        // ============================================================
        // 5. PUTTING IT ALL TOGETHER — Mini Spring Container
        // ============================================================
        System.out.println("\n=== 5. MINI SPRING CONTAINER ===");
        System.out.println("(What Spring does at startup)\n");

        // Step 1: Spring finds the class name (from component scanning)
        String className = "PaymentService";
        Class<?> foundClass = Class.forName(className);
        System.out.println("Step 1: Found class → " + foundClass.getSimpleName());

        // Step 2: Spring creates an instance (a "bean")
        Object bean = foundClass.getDeclaredConstructor().newInstance();
        System.out.println("Step 2: Created bean → " + bean);

        // Step 3: Spring injects a value into a private field
        Field f = foundClass.getDeclaredField("provider");
        f.setAccessible(true);
        f.set(bean, "GooglePay");
        System.out.println("Step 3: Injected provider → " + ((PaymentService) bean).getProvider());

        // Step 4: Spring invokes a method
        Method m = foundClass.getDeclaredMethod("processPayment", double.class);
        System.out.print("Step 4: Invoked → ");
        m.invoke(bean, 99.99);

        System.out.println("\n=== KEY TAKEAWAY ===");
        System.out.println("Reflection = Java's ability to inspect & manipulate code at runtime.");
        System.out.println("Spring uses reflection to CREATE beans, INJECT dependencies, READ annotations.");
        System.out.println("Now you know there's no magic — just reflection + annotations.");
    }
}
