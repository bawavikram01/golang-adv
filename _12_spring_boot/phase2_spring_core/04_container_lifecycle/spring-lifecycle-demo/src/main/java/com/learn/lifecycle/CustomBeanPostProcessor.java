package com.learn.lifecycle;

import org.springframework.beans.BeansException;
import org.springframework.beans.factory.config.BeanPostProcessor;
import org.springframework.stereotype.Component;

/**
 * A BeanPostProcessor runs for EVERY bean created by the container.
 * It intercepts beans before and after initialization.
 *
 * Real-world uses:
 * - @Autowired processing (AutowiredAnnotationBeanPostProcessor)
 * - AOP proxy creation
 * - Custom annotation processing
 * - Logging/metrics around bean initialization
 */
@Component
public class CustomBeanPostProcessor implements BeanPostProcessor {

    @Override
    public Object postProcessBeforeInitialization(Object bean, String beanName) throws BeansException {
        // Only log for our demo bean to keep output clean
        if (bean instanceof FullLifecycleBean) {
            System.out.println("  [Step  4] BeanPostProcessor.postProcessBEFOREInitialization() → " + beanName);
        }
        return bean; // Must return the bean (or a wrapper/proxy)
    }

    @Override
    public Object postProcessAfterInitialization(Object bean, String beanName) throws BeansException {
        // Only log for our demo bean
        if (bean instanceof FullLifecycleBean) {
            System.out.println("  [Step  8] BeanPostProcessor.postProcessAFTERInitialization() → " + beanName);
            System.out.println("            (AOP proxies would be created HERE)");
        }
        return bean; // In real AOP, this would return a proxy wrapping the bean
    }
}
