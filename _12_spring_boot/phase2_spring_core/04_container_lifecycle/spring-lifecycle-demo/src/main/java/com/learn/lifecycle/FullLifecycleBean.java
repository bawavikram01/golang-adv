package com.learn.lifecycle;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.BeanFactory;
import org.springframework.beans.factory.BeanFactoryAware;
import org.springframework.beans.factory.BeanNameAware;
import org.springframework.beans.factory.DisposableBean;
import org.springframework.beans.factory.InitializingBean;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.stereotype.Component;

/**
 * This bean implements ALL lifecycle interfaces to demonstrate
 * the exact order Spring calls each method.
 *
 * Real-world: You'd typically only use @PostConstruct and @PreDestroy.
 * We implement everything here purely for educational purposes.
 */
@Component
public class FullLifecycleBean implements
        BeanNameAware,
        BeanFactoryAware,
        ApplicationContextAware,
        InitializingBean,
        DisposableBean {

    private String beanName;
    private int step = 0;

    // ─────────────────────────────────────────────────────────────
    // STEP 1: CONSTRUCTOR — Bean is instantiated
    // ─────────────────────────────────────────────────────────────
    public FullLifecycleBean() {
        print("CONSTRUCTOR called — bean instantiated (no dependencies yet)");
    }

    // ─────────────────────────────────────────────────────────────
    // STEP 2: DEPENDENCY INJECTION
    // (Happens automatically if we had @Autowired fields/setters)
    // ─────────────────────────────────────────────────────────────

    // ─────────────────────────────────────────────────────────────
    // STEP 3: AWARE INTERFACES — Spring tells bean about itself
    // ─────────────────────────────────────────────────────────────
    @Override
    public void setBeanName(String name) {
        this.beanName = name;
        print("BeanNameAware.setBeanName() → name = \"" + name + "\"");
    }

    @Override
    public void setBeanFactory(BeanFactory beanFactory) throws BeansException {
        print("BeanFactoryAware.setBeanFactory() → got reference to BeanFactory");
    }

    @Override
    public void setApplicationContext(ApplicationContext ctx) throws BeansException {
        print("ApplicationContextAware.setApplicationContext() → got full container");
    }

    // ─────────────────────────────────────────────────────────────
    // STEP 4: BeanPostProcessor.postProcessBeforeInitialization()
    // (See CustomBeanPostProcessor.java — runs for ALL beans)
    // ─────────────────────────────────────────────────────────────

    // ─────────────────────────────────────────────────────────────
    // STEP 5: INITIALIZATION PHASE (3 options, all called in order)
    // ─────────────────────────────────────────────────────────────
    @PostConstruct
    public void postConstruct() {
        print("@PostConstruct — initialization logic (MOST COMMON hook)");
    }

    @Override
    public void afterPropertiesSet() throws Exception {
        print("InitializingBean.afterPropertiesSet() — older style init");
    }

    // Note: Custom init-method would be defined via @Bean(initMethod="...")
    // We skip that here as @PostConstruct is the modern replacement.

    // ─────────────────────────────────────────────────────────────
    // STEP 6: BeanPostProcessor.postProcessAfterInitialization()
    // (See CustomBeanPostProcessor.java — AOP proxies happen here)
    // ─────────────────────────────────────────────────────────────

    // ─────────────────────────────────────────────────────────────
    // STEP 7: BEAN IS READY — Application uses it
    // ─────────────────────────────────────────────────────────────
    public void doWork() {
        System.out.println("\n  ✓ Bean \"" + beanName + "\" doing actual work! (fully initialized)\n");
    }

    // ─────────────────────────────────────────────────────────────
    // STEP 8: DESTRUCTION (on application shutdown)
    // ─────────────────────────────────────────────────────────────
    @PreDestroy
    public void preDestroy() {
        print("@PreDestroy — cleanup resources (MOST COMMON destroy hook)");
    }

    @Override
    public void destroy() throws Exception {
        print("DisposableBean.destroy() — older style destroy");
    }

    // ─────────────────────────────────────────────────────────────
    // Helper
    // ─────────────────────────────────────────────────────────────
    private void print(String message) {
        step++;
        System.out.printf("  [Step %2d] %s%n", step, message);
    }
}
