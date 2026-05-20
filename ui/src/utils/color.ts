/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import Color from 'color';

/**
 * Bootstrap Color Weight:
 * $blue-100: tint-color($blue, 80%) !default;
 * $blue-200: tint-color($blue, 60%) !default;
 * $blue-300: tint-color($blue, 40%) !default;
 * $blue-400: tint-color($blue, 20%) !default;
 * $blue-500: $blue !default;
 * $blue-600: shade-color($blue, 20%) !default;
 * $blue-700: shade-color($blue, 40%) !default;
 * $blue-800: shade-color($blue, 60%) !default;
 * $blue-900: shade-color($blue, 80%) !default;
 */

/**
 *  The `weight` parameter in `Color`:
 *    1. Must use decimals rather than percentages. eg: color.mix(Color("blue"), 0.6)
 *    2. The value is the difference between `1 - $weight` in `bootstrap`.
 *      eg: color.mix(Color("blue"), 0.6) === shade-color($blue, 40%) !default
 */

const WHITE = Color('#fff');
const BLACK = Color('#000');

export const mixColour = (baseColor, opColor, weight) => {
  return Color(baseColor).mix(Color(opColor), weight);
};

export const tintColor = (color, weight) => {
  return mixColour(WHITE, color, weight);
};

export const shadeColor = (color, weight) => {
  return mixColour(BLACK, color, weight);
};

export const shiftColor = (color, weight) => {
  if (weight > 0) {
    return shadeColor(color, weight);
  }
  return tintColor(color, -weight);
};

export const isLight = (color) => {
  return Color(color).isLight();
};
