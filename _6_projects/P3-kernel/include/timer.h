#pragma once
#include <stdint.h>

//
// timer.h — Programmable Interval Timer (PIT) driver
//

void timer_init(uint32_t frequency_hz);
uint32_t timer_get_ticks();
