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

import { RefObject } from 'react';

import builtin from '@/plugins/builtin';
import * as allPlugins from '@/plugins';
import type * as Type from '@/common/interface';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';
import { getPluginsStatus } from '@/services';
import Storage from '@/utils/storage';
import request from '@/utils/request';

import { initI18nResource } from './utils';
import { Plugin, PluginInfo, PluginType } from './interface';

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

class Plugins {
  plugins: Plugin[] = [];

  registeredPlugins: Type.ActivatedPlugin[] = [];

  initialization: Promise<void>;

  private isInitialized = false;

  private initializationError: Error | null = null;

  private replacementPlugins: Map<PluginType, Plugin> = new Map();

  constructor() {
    this.initialization = this.init();
  }

  async init() {
    if (this.isInitialized) {
      return;
    }

    try {
      this.registerBuiltin();

      // Note: The /install stage does not allow access to the getPluginsStatus api, so an initial value needs to be given
      const plugins =
        (await getPluginsStatus().catch((error) => {
          console.warn('Failed to get plugins status:', error);
          return [];
        })) || [];
      this.registeredPlugins = plugins.filter((p) => p.enabled);
      await this.registerPlugins();
      this.isInitialized = true;
      this.initializationError = null;
    } catch (error) {
      this.initializationError = error as Error;
      console.error('Plugin initialization failed:', error);
      throw error;
    }
  }

  async refresh() {
    this.plugins = [];
    this.isInitialized = false;
    this.initializationError = null;
    this.initialization = this.init();
    await this.initialization;
  }

  validate(plugin: Plugin) {
    if (!plugin) {
      return false;
    }
    const { info } = plugin;
    const { slug_name, type } = info;

    if (!slug_name) {
      return false;
    }

    if (!type) {
      return false;
    }

    return true;
  }

  registerBuiltin() {
    Object.keys(builtin).forEach((key) => {
      const plugin = builtin[key];
      // builttin plugins are always activated
      // Use own internal rendering logic'
      plugin.activated = true;
      this.register(plugin);
    });
  }

  async registerPlugins() {
    console.log(
      '[PluginKit] Registered plugins from API:',
      this.registeredPlugins.map((p) => p.slug_name),
    );

    const pluginLoaders = this.registeredPlugins
      .map((p) => {
        const func = allPlugins[p.slug_name];
        if (!func) {
          console.warn(
            `[PluginKit] Plugin loader not found for: ${p.slug_name}`,
          );
        }
        return { slug_name: p.slug_name, loader: func };
      })
      .filter((p) => p.loader);

    console.log(
      '[PluginKit] Found plugin loaders:',
      pluginLoaders.map((p) => p.slug_name),
    );

    // Use Promise.allSettled to prevent one plugin failure from breaking all plugins
    const results = await Promise.allSettled(
      pluginLoaders.map((p) => p.loader()),
    );

    results.forEach((result, index) => {
      if (result.status === 'fulfilled') {
        console.log(
          `[PluginKit] Successfully loaded plugin: ${pluginLoaders[index].slug_name}`,
        );
        this.register(result.value);
      } else {
        console.error(
          `[PluginKit] Failed to load plugin ${pluginLoaders[index].slug_name}:`,
          result.reason,
        );
      }
    });
  }

  register(plugin: Plugin) {
    const bool = this.validate(plugin);
    if (!bool) {
      return;
    }

    // Prevent duplicate registration
    const exists = this.plugins.some(
      (p) => p.info.slug_name === plugin.info.slug_name,
    );
    if (exists) {
      console.warn(`Plugin ${plugin.info.slug_name} is already registered`);
      return;
    }

    // Handle singleton plugins (only one per type allowed)
    const mode = plugin.info.registrationMode || 'multiple';
    if (mode === 'singleton') {
      const existingPlugin = this.replacementPlugins.get(plugin.info.type);
      if (existingPlugin) {
        const error = new Error(
          `[PluginKit] Plugin conflict: ` +
            `Cannot register '${plugin.info.slug_name}' because '${existingPlugin.info.slug_name}' ` +
            `is already registered as a singleton plugin of type '${plugin.info.type}'. ` +
            `Only one singleton plugin per type is allowed.`,
        );
        console.error(error.message);
        throw error;
      }
      this.replacementPlugins.set(plugin.info.type, plugin);
    }

    if (plugin.i18nConfig) {
      initI18nResource(plugin.i18nConfig);
    }
    plugin.activated = true;
    this.plugins.push(plugin);
  }

  getPlugin(slug_name: string) {
    return this.plugins.find((p) => p.info.slug_name === slug_name);
  }

  getOnePluginHooks(slug_name: string) {
    const plugin = this.getPlugin(slug_name);
    return plugin?.hooks;
  }

  getPlugins() {
    return this.plugins;
  }

  async getPluginsAsync() {
    await this.initialization;
    return this.plugins;
  }

  getInitializationStatus() {
    return {
      isInitialized: this.isInitialized,
      error: this.initializationError,
    };
  }

  getReplacementPlugin(type: PluginType): Plugin | null {
    return this.replacementPlugins.get(type) || null;
  }
}

const plugins = new Plugins();

const getRoutePlugins = async () => {
  await plugins.initialization;
  return plugins
    .getPlugins()
    .filter((plugin) => plugin.info.type === PluginType.Route);
};

const defaultProps = () => {
  const token = Storage.get(LOGGED_TOKEN_STORAGE_KEY) || '';
  return {
    key: token,
    headers: {
      Authorization: token,
    },
  };
};

const validateRoutePlugin = async (slugName) => {
  let registeredPlugin;
  if (plugins.registeredPlugins.length === 0) {
    const pluginsStatus = await getPluginsStatus();
    registeredPlugin = pluginsStatus.find((p) => p.slug_name === slugName);
  } else {
    registeredPlugin = plugins.registeredPlugins.find(
      (p) => p.slug_name === slugName,
    );
  }

  return Boolean(registeredPlugin?.enabled);
};

const getReplacementPlugin = async (
  type: PluginType,
): Promise<Plugin | null> => {
  try {
    await plugins.initialization;
    return plugins.getReplacementPlugin(type);
  } catch (error) {
    console.error(
      `[PluginKit] Failed to get replacement plugin of type ${type}:`,
      error,
    );
    return null;
  }
};

const mergeRoutePlugins = async (routes) => {
  const routePlugins = await getRoutePlugins();
  if (routePlugins.length === 0) {
    return routes;
  }
  routes.forEach((route) => {
    if (route.page === 'pages/Layout') {
      route.children?.forEach((child) => {
        if (child.page === 'pages/SideNavLayout') {
          routePlugins.forEach((plugin) => {
            const { route: path, slug_name } = plugin.info;
            child.children.push({
              page: plugin.component,
              path,
              loader: async () => {
                const bool = await validateRoutePlugin(slug_name);
                return bool;
              },
              guard: (params) => {
                if (params.loaderData) {
                  return {
                    ok: true,
                  };
                }

                return {
                  ok: false,
                  error: {
                    code: 404,
                  },
                };
              },
            });
          });
        }
      });
    }
  });
  return routes;
};

/**
 * Only used to enhance the capabilities of the markdown editor
 * Add RefObject type to solve the problem of dom being null in hooks
 */
const useRenderHtmlPlugin = (
  element: HTMLElement | RefObject<HTMLElement> | null,
) => {
  plugins
    .getPlugins()
    .filter((plugin) => {
      return (
        plugin.activated &&
        plugin.hooks?.useRender &&
        (plugin.info.type === PluginType.Editor ||
          plugin.info.type === PluginType.Render)
      );
    })
    .forEach((plugin) => {
      plugin.hooks?.useRender?.forEach((hook) => {
        hook(element, request);
      });
    });
};

// Only for render type plugins
const useRenderPlugin = (
  element: HTMLElement | RefObject<HTMLElement> | null,
) => {
  return plugins
    .getPlugins()
    .filter((plugin) => {
      return (
        plugin.activated &&
        plugin.hooks?.useRender &&
        plugin.info.type === PluginType.Render
      );
    })
    .forEach((plugin) => {
      plugin.hooks?.useRender?.forEach((hook) => {
        hook(element, request);
      });
    });
};

// Only one captcha type plug-in can be enabled at the same time
const useCaptchaPlugin = (key: Type.CaptchaKey) => {
  const captcha = plugins
    .getPlugins()
    .filter(
      (plugin) => plugin.info.type === PluginType.Captcha && plugin.activated,
    );
  const pluginHooks = plugins.getOnePluginHooks(captcha[0]?.info.slug_name);
  return pluginHooks?.useCaptcha?.({
    captchaKey: key,
    commonProps: defaultProps(),
  });
};

export type { Plugin, PluginInfo };

export {
  useRenderHtmlPlugin,
  mergeRoutePlugins,
  useCaptchaPlugin,
  useRenderPlugin,
  getReplacementPlugin,
  PluginType,
};
export default plugins;
