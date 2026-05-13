/*
 * ============================================================
 *  CHAPTER 25: SERIALIZATION
 * ============================================================
 *  Converting objects to byte streams (and back) for:
 *  - Saving to disk
 *  - Sending over network
 *  - Caching
 *
 *  Serializable   → marker interface, automatic serialization
 *  Externalizable → full manual control
 *  transient      → field is excluded from serialization
 *  serialVersionUID → version control for serialized classes
 * ============================================================
 */

import java.io.*;

public class Chapter25_Serialization {

    // --- Serializable class ---
    static class Person implements Serializable {
        private static final long serialVersionUID = 1L; // version control

        String name;
        int age;
        transient String password; // NOT serialized!

        Person(String name, int age, String password) {
            this.name = name;
            this.age = age;
            this.password = password;
        }

        @Override
        public String toString() {
            return "Person{name='" + name + "', age=" + age + ", password='" + password + "'}";
        }
    }

    // --- Custom serialization ---
    static class SecureUser implements Serializable {
        private static final long serialVersionUID = 2L;
        String username;
        transient String secret;

        SecureUser(String username, String secret) {
            this.username = username;
            this.secret = secret;
        }

        // Custom serialization hooks
        private void writeObject(ObjectOutputStream out) throws IOException {
            out.defaultWriteObject();
            // Write encrypted version of secret
            out.writeObject("ENC:" + new StringBuilder(secret).reverse());
        }

        private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
            in.defaultReadObject();
            // Decrypt the secret
            String encrypted = (String) in.readObject();
            this.secret = new StringBuilder(encrypted.substring(4)).reverse().toString();
        }

        @Override
        public String toString() {
            return "SecureUser{username='" + username + "', secret='" + secret + "'}";
        }
    }

    public static void main(String[] args) {

        String filename = "person.ser";

        // --- 1. Serialize (Save) ---
        System.out.println("=== SERIALIZATION ===\n");
        Person person = new Person("Alice", 25, "secret123");
        System.out.println("Before: " + person);

        try (ObjectOutputStream oos = new ObjectOutputStream(new FileOutputStream(filename))) {
            oos.writeObject(person);
            System.out.println("Serialized to: " + filename);
        } catch (IOException e) {
            System.out.println("Error: " + e.getMessage());
        }

        // --- 2. Deserialize (Load) ---
        System.out.println("\n=== DESERIALIZATION ===\n");
        try (ObjectInputStream ois = new ObjectInputStream(new FileInputStream(filename))) {
            Person loaded = (Person) ois.readObject();
            System.out.println("After: " + loaded);
            System.out.println("Notice: password is null (transient field)!");
        } catch (IOException | ClassNotFoundException e) {
            System.out.println("Error: " + e.getMessage());
        }

        // --- 3. Custom serialization ---
        System.out.println("\n=== CUSTOM SERIALIZATION ===\n");
        String secFile = "secure.ser";
        SecureUser user = new SecureUser("admin", "mypassword");
        System.out.println("Before: " + user);

        try (ObjectOutputStream oos = new ObjectOutputStream(new FileOutputStream(secFile))) {
            oos.writeObject(user);
        } catch (IOException e) { e.printStackTrace(); }

        try (ObjectInputStream ois = new ObjectInputStream(new FileInputStream(secFile))) {
            SecureUser loaded = (SecureUser) ois.readObject();
            System.out.println("After: " + loaded);
        } catch (IOException | ClassNotFoundException e) { e.printStackTrace(); }

        // Cleanup
        new File(filename).delete();
        new File(secFile).delete();

        // --- Summary ---
        System.out.println("\n=== KEY POINTS ===");
        System.out.println("1. Implement Serializable to enable serialization");
        System.out.println("2. Use transient for sensitive/non-serializable fields");
        System.out.println("3. Always define serialVersionUID for version control");
        System.out.println("4. Use writeObject/readObject for custom logic");
        System.out.println("5. Consider JSON/XML for cross-platform serialization");
    }
}

/*
 * NEXT: Chapter 26 — Date & Time API
 */
