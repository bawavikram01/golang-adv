/*
 * ============================================================
 *  CHAPTER 41: NETWORKING & SOCKETS
 * ============================================================
 *  Java has powerful networking built-in:
 *    java.net   — Sockets, URLs, HTTP
 *    java.nio   — Non-blocking I/O (NIO)
 *
 *  TOPICS:
 *    1. TCP Client/Server (Socket / ServerSocket)
 *    2. UDP (DatagramSocket)
 *    3. HTTP requests (HttpURLConnection, HttpClient)
 *    4. URL parsing
 * ============================================================
 */

import java.net.*;
import java.io.*;
import java.util.concurrent.*;

public class Chapter41_Networking {

    // ========================================================
    // 1. TCP SERVER (runs in background thread)
    // ========================================================
    static class SimpleTCPServer implements Runnable {
        private final int port;
        private volatile boolean running = true;

        SimpleTCPServer(int port) { this.port = port; }

        @Override
        public void run() {
            try (ServerSocket server = new ServerSocket(port)) {
                server.setSoTimeout(5000);  // timeout for accept()
                System.out.println("  [Server] Listening on port " + port);

                while (running) {
                    try {
                        Socket client = server.accept();
                        handleClient(client);
                    } catch (SocketTimeoutException e) {
                        // timeout, check if still running
                    }
                }
            } catch (IOException e) {
                System.out.println("  [Server] Error: " + e.getMessage());
            }
        }

        private void handleClient(Socket client) throws IOException {
            try (
                BufferedReader in = new BufferedReader(new InputStreamReader(client.getInputStream()));
                PrintWriter out = new PrintWriter(client.getOutputStream(), true)
            ) {
                String message = in.readLine();
                System.out.println("  [Server] Received: " + message);

                // Echo back uppercase
                out.println("ECHO: " + message.toUpperCase());
                System.out.println("  [Server] Sent response");
            } finally {
                client.close();
            }
        }

        void stop() { running = false; }
    }

    // ========================================================
    // 2. TCP CLIENT
    // ========================================================
    static String sendTCPMessage(String host, int port, String message) throws IOException {
        try (
            Socket socket = new Socket(host, port);
            PrintWriter out = new PrintWriter(socket.getOutputStream(), true);
            BufferedReader in = new BufferedReader(new InputStreamReader(socket.getInputStream()))
        ) {
            out.println(message);
            return in.readLine();
        }
    }

    // ========================================================
    // 3. MULTI-THREADED SERVER (handles multiple clients)
    // ========================================================
    static class MultiThreadedServer implements Runnable {
        private final int port;
        private volatile boolean running = true;

        MultiThreadedServer(int port) { this.port = port; }

        @Override
        public void run() {
            ExecutorService pool = Executors.newFixedThreadPool(10);
            try (ServerSocket server = new ServerSocket(port)) {
                server.setSoTimeout(3000);
                System.out.println("  [MT Server] Listening on port " + port);

                while (running) {
                    try {
                        Socket client = server.accept();
                        pool.submit(() -> {
                            try (
                                BufferedReader in = new BufferedReader(
                                    new InputStreamReader(client.getInputStream()));
                                PrintWriter out = new PrintWriter(client.getOutputStream(), true)
                            ) {
                                String msg = in.readLine();
                                out.println("MT-ECHO: " + msg);
                            } catch (IOException e) {
                                // handle error
                            } finally {
                                try { client.close(); } catch (IOException e) {}
                            }
                        });
                    } catch (SocketTimeoutException e) {
                        // check running flag
                    }
                }
            } catch (IOException e) {
                System.out.println("  [MT Server] Error: " + e.getMessage());
            } finally {
                pool.shutdown();
            }
        }

        void stop() { running = false; }
    }

    // ========================================================
    // 4. UDP (Connectionless)
    // ========================================================
    static class UDPServer implements Runnable {
        private final int port;
        private volatile boolean running = true;

        UDPServer(int port) { this.port = port; }

        @Override
        public void run() {
            try (DatagramSocket socket = new DatagramSocket(port)) {
                socket.setSoTimeout(3000);
                System.out.println("  [UDP Server] Listening on port " + port);
                byte[] buffer = new byte[1024];

                while (running) {
                    try {
                        DatagramPacket packet = new DatagramPacket(buffer, buffer.length);
                        socket.receive(packet);
                        String message = new String(packet.getData(), 0, packet.getLength());
                        System.out.println("  [UDP Server] Received: " + message);

                        // Send response
                        byte[] response = ("UDP-ECHO: " + message).getBytes();
                        DatagramPacket reply = new DatagramPacket(
                            response, response.length, packet.getAddress(), packet.getPort());
                        socket.send(reply);
                    } catch (SocketTimeoutException e) {
                        // check running flag
                    }
                }
            } catch (IOException e) {
                System.out.println("  [UDP Server] Error: " + e.getMessage());
            }
        }

        void stop() { running = false; }
    }

    static String sendUDP(String host, int port, String message) throws IOException {
        try (DatagramSocket socket = new DatagramSocket()) {
            socket.setSoTimeout(3000);
            byte[] data = message.getBytes();
            DatagramPacket packet = new DatagramPacket(
                data, data.length, InetAddress.getByName(host), port);
            socket.send(packet);

            byte[] buffer = new byte[1024];
            DatagramPacket response = new DatagramPacket(buffer, buffer.length);
            socket.receive(response);
            return new String(response.getData(), 0, response.getLength());
        }
    }

    public static void main(String[] args) throws Exception {

        // --- 1. URL Parsing ---
        System.out.println("=== URL PARSING ===\n");
        URL url = new URL("https://www.example.com:443/path/page?query=java#section");
        System.out.println("  Protocol: " + url.getProtocol());
        System.out.println("  Host: " + url.getHost());
        System.out.println("  Port: " + url.getPort());
        System.out.println("  Path: " + url.getPath());
        System.out.println("  Query: " + url.getQuery());
        System.out.println("  Ref: " + url.getRef());

        // --- 2. InetAddress ---
        System.out.println("\n=== INET ADDRESS ===\n");
        InetAddress local = InetAddress.getLocalHost();
        System.out.println("  Local hostname: " + local.getHostName());
        System.out.println("  Local IP: " + local.getHostAddress());

        InetAddress loopback = InetAddress.getLoopbackAddress();
        System.out.println("  Loopback: " + loopback.getHostAddress());

        // --- 3. TCP Client-Server Demo ---
        System.out.println("\n=== TCP CLIENT-SERVER ===\n");
        int tcpPort = 9876;
        SimpleTCPServer tcpServer = new SimpleTCPServer(tcpPort);
        Thread serverThread = new Thread(tcpServer);
        serverThread.start();
        Thread.sleep(500);  // let server start

        String response = sendTCPMessage("localhost", tcpPort, "Hello Java Networking!");
        System.out.println("  [Client] Response: " + response);

        tcpServer.stop();
        serverThread.join(5000);

        // --- 4. UDP Demo ---
        System.out.println("\n=== UDP CLIENT-SERVER ===\n");
        int udpPort = 9877;
        UDPServer udpServer = new UDPServer(udpPort);
        Thread udpThread = new Thread(udpServer);
        udpThread.start();
        Thread.sleep(500);

        String udpResponse = sendUDP("localhost", udpPort, "Hello UDP!");
        System.out.println("  [Client] Response: " + udpResponse);

        udpServer.stop();
        udpThread.join(5000);

        // --- 5. TCP vs UDP ---
        System.out.println("\n=== TCP vs UDP ===");
        System.out.println("  TCP                          UDP");
        System.out.println("  ───                          ───");
        System.out.println("  Connection-oriented          Connectionless");
        System.out.println("  Reliable (guaranteed)        Unreliable (best effort)");
        System.out.println("  Ordered delivery             No ordering");
        System.out.println("  Slower (overhead)            Faster (lightweight)");
        System.out.println("  HTTP, FTP, SSH, Email        DNS, Video, Gaming, VoIP");

        // --- 6. HTTP with HttpURLConnection ---
        System.out.println("\n=== HTTP CONCEPTS ===");
        System.out.println("  GET    — retrieve data");
        System.out.println("  POST   — submit data");
        System.out.println("  PUT    — update/replace");
        System.out.println("  DELETE — remove");
        System.out.println("  PATCH  — partial update");
        System.out.println();
        System.out.println("  Status codes:");
        System.out.println("  200 OK, 201 Created, 204 No Content");
        System.out.println("  301 Moved, 400 Bad Request, 401 Unauthorized");
        System.out.println("  403 Forbidden, 404 Not Found, 500 Server Error");

        System.out.println("\n✓ Networking Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Build a chat server: multiple clients send messages, server broadcasts to all.
 * 2. Create a simple HTTP server that serves HTML files.
 * 3. Build a file transfer tool using TCP sockets.
 * 4. Use HttpClient (Java 11+) to call a REST API and parse the JSON response.
 *
 * NEXT: Chapter 42 — JDBC & Database
 */
