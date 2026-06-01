package com.learn.rest.service;

import com.learn.rest.model.User;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.atomic.AtomicLong;

/**
 * In-memory "database" using a HashMap.
 * In real apps, this would be replaced by a JPA Repository (Phase 5).
 */
@Service
public class UserService {

    private final Map<Long, User> users = new LinkedHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(1);

    // Pre-populate some data
    public UserService() {
        save(new User(null, "Alice Johnson", "alice@example.com"));
        save(new User(null, "Bob Smith", "bob@example.com"));
        save(new User(null, "Charlie Brown", "charlie@example.com"));
    }

    public List<User> findAll() {
        return new ArrayList<>(users.values());
    }

    public Optional<User> findById(Long id) {
        return Optional.ofNullable(users.get(id));
    }

    public List<User> findByName(String name) {
        return users.values().stream()
                .filter(u -> u.getName().toLowerCase().contains(name.toLowerCase()))
                .toList();
    }

    public User save(User user) {
        if (user.getId() == null) {
            user.setId(idGenerator.getAndIncrement());
        }
        users.put(user.getId(), user);
        return user;
    }

    public Optional<User> update(Long id, User updated) {
        if (!users.containsKey(id)) {
            return Optional.empty();
        }
        updated.setId(id);
        users.put(id, updated);
        return Optional.of(updated);
    }

    public boolean delete(Long id) {
        return users.remove(id) != null;
    }
}
