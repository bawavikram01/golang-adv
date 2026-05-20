package com.learn.profiles;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

/**
 * DEV: In-memory H2 database — fast, disposable, no setup needed.
 */
@Component
@Profile({"dev", "default"})  // Active in dev OR when no profile set
public class H2DataSourceService implements DataSourceService {

    @Override
    public String getInfo() {
        return "H2 In-Memory DB (jdbc:h2:mem:devdb) — fast, ephemeral";
    }
}
