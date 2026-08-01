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

import i18next from 'i18next';

import { PluginInfo, Plugin, PluginType } from './interface';

/**
 * This information is to be defined for all components.
 * It may be used for feature upgrades or version compatibility processing.
 *
 * @field slug_name: Unique identity string for the plugin, usually configured in `info.yaml`
 * @field type: The type of plugin is defined and a single type of plugin can have multiple implementations.
 *              For example, a plugin of type `connector` can have a `google` implementation and a `github` implementation.
 *              `PluginRender` automatically renders the plug-in types already included in `PluginType`.
 * @field name: Plugin name, optionally configurable. Usually read from the `i18n` file
 * @field description: Plugin description, optionally configurable. Usually read from the `i18n` file
 */

const I18N_NS = 'plugin';
interface I18nResource {
  [lng: string]: {
    plugin: {
      [slug_name: string]: {
        ui: any;
      };
    };
  };
}

const addResourceBundle = (resource: I18nResource) => {
  if (resource) {
    Object.keys(resource).forEach((lng) => {
      const r = resource[lng];

      i18next.addResourceBundle(lng, I18N_NS, r.plugin, true, true);
    });
  }
};

const initI18nResource = (resource: I18nResource) => {
  /**
   * `i18next` only attaches its resource-store methods to the instance while
   * `init` runs, so they do not exist yet if a plugin's i18n module happens to
   * evaluate first. Module evaluation order is decided by the bundler, so
   * registering unconditionally works or crashes depending on how the chunks
   * come out. Register now only when there is something to register into, and
   * otherwise let the handler below do it.
   */
  if (i18next.isInitialized) {
    addResourceBundle(resource);
  }
  /**
   * Note: In development mode,
   * when the base i18n file is changed, `i18next` will reinitialise the updated resource file,
   * which will cause the resource package added in the plugin to be lost
   * and will need to be automatically re-added by listening for events
   */
  i18next.on('initialized', () => {
    addResourceBundle(resource);
  });
};

const getTransNs = () => {
  return I18N_NS;
};

const getTransKeyPrefix = (info: PluginInfo) => {
  const kp = `${info.slug_name}.ui`;
  return kp;
};

export type { Plugin, PluginInfo, PluginType };

export { initI18nResource, getTransNs, getTransKeyPrefix };
