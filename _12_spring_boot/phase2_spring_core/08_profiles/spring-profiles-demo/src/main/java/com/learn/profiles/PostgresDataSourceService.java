package com.learn.profiles;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

/**
 * PROD: PostgreSQL connection pool — real persistent database.
 */
@Component
@Profile("prod")
public class PostgresDataSourceService implements DataSourceService {

    @Override
    public String getInfo() {
        return "PostgreSQL (jdbc:postgresql://prod-server:5432/myapp) — persistent, pooled";
    }
}
