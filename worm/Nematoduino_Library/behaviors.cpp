#include <stdlib.h>
#include <vector>

#include "behaviors.hpp"

// Arrays for neurons describing 'nose touch' and
// food-seeking behaviors
const uint16_t NOSE_TOUCH[] = {
  N_FLPR, N_FLPL, N_ASHL, N_ASHR, N_IL1VL, N_IL1VR,
  N_OLQDL, N_OLQDR, N_OLQVR, N_OLQVL
};

const uint16_t NOSE_TOUCH_FR[] = {
  N_FLPR
};

const uint16_t NOSE_TOUCH_FL[] = {
  N_FLPL
};

const uint16_t NOSE_TOUCH_AL[] = {
  N_ASHL
};

const uint16_t NOSE_TOUCH_AR[] = {
  N_ASHR
};

const uint16_t NOSE_TOUCH_VL[] = {
  N_IL1VL
};

const uint16_t NOSE_TOUCH_VR[] = {
  N_IL1VR
};

const uint16_t NOSE_TOUCH_OL[] = {
  N_OLQDL
};

const uint16_t NOSE_TOUCH_OR[] = {
  N_OLQDR
};

const uint16_t NOSE_TOUCH_OVL[] = {
  N_OLQVL
};

const uint16_t NOSE_TOUCH_OVR[] = {
  N_OLQVR
};

const uint16_t CHEMOTAXIS[] = {
  N_ADFL, N_ADFR, N_ASGR, N_ASGL, N_ASIL, N_ASIR,
  N_ASJR, N_ASJL
};