package com.learn.springdemo;

import org.springframework.stereotype.Component;

/**
 * A simple config bean.
 * @Component tells Spring: "Create me, store me, I'm a bean."
 */
@Component
public class AppConfig {

    public String getDbUrl() {
        return "jdbc:h2:mem:myapp";
    }

    public String getAppName() {
        return "SpringDemo";
    }
}
