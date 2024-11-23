#ifndef BEHAVIORS_H
#define BEHAVIORS_H

#define CHEMOTAXIS_LEN 8
#define NOSE_TOUCH_LEN 10

#include <stdint.h>

extern "C" {
#include "utility/defines.h"
}

extern const uint16_t NOSE_TOUCH[];
extern const uint16_t NOSE_TOUCH_FR[];
extern const uint16_t NOSE_TOUCH_FL[];
extern const uint16_t NOSE_TOUCH_AL[];
extern const uint16_t NOSE_TOUCH_AR[];
extern const uint16_t NOSE_TOUCH_VL[];
extern const uint16_t NOSE_TOUCH_VR[];
extern const uint16_t NOSE_TOUCH_OL[];
extern const uint16_t NOSE_TOUCH_OR[];
extern const uint16_t NOSE_TOUCH_OVL[];
extern const uint16_t NOSE_TOUCH_OVR[];



extern const uint16_t CHEMOTAXIS[];

#endif