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

import React, { FC, ReactNode, useEffect, useState } from 'react';

import PluginKit, { Plugin, PluginType } from '@/utils/pluginKit';

// Marker component for plugin insertion point
export const PluginSlot: FC = () => null;
/**
 * Note：Please set at least either of the `slug_name` and `type` attributes, otherwise no plugins will be rendered.
 *
 * @field slug_name: The `slug_name` of the plugin needs to be rendered.
 *                   If this property is set, `PluginRender` will use it first (regardless of whether `type` is set)
 *                   to find the corresponding plugin and render it.
 * @field type: Used to formulate the rendering of all plugins of this type.
 *              (if the `slug_name` attribute is set, it will be ignored)
 * @field prop: Any attribute you want to configure, e.g. `className`
 *
 * For editor type plugins, use <PluginSlot /> component as a marker to indicate where plugins should be inserted.
 */

interface Props {
  slug_name?: string;
  type: PluginType;
  children?: ReactNode;
  className?: string;
  [key: string]: unknown;
}

const Index: FC<Props> = ({
  slug_name,
  type,
  children = null,
  className,
  ...props
}) => {
  const [pluginSlice, setPluginSlice] = useState<Plugin[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let mounted = true;

    const loadPlugins = async () => {
      await PluginKit.initialization;

      if (!mounted) return;

      const plugins = PluginKit.getPlugins().filter(
        (plugin) => plugin.activated,
      );
      console.log(
        '[PluginRender] Loaded plugins:',
        plugins.map((p) => p.info.slug_name),
      );
      const filtered: Plugin[] = [];

      plugins.forEach((plugin) => {
        if (type && slug_name) {
          if (
            plugin.info.slug_name === slug_name &&
            plugin.info.type === type
          ) {
            filtered.push(plugin);
          }
        } else if (type) {
          if (plugin.info.type === type) {
            filtered.push(plugin);
          }
        } else if (slug_name) {
          if (plugin.info.slug_name === slug_name) {
            filtered.push(plugin);
          }
        }
      });

      if (mounted) {
        setPluginSlice(filtered);
        setIsLoading(false);
      }
    };

    loadPlugins();

    return () => {
      mounted = false;
    };
  }, [slug_name, type]);

  /**
   * TODO: Rendering control for non-builtin plug-ins
   * ps: Logic such as version compatibility determination can be placed here
   */
  if (isLoading) {
    // Don't render anything while loading to avoid flashing
    if (type === 'editor') {
      return <div className={className}>{children}</div>;
    }
    return null;
  }

  if (pluginSlice.length === 0) {
    if (type === 'editor') {
      return <div className={className}>{children}</div>;
    }
    return null;
  }

  if (type === 'editor') {
    // Use PluginSlot marker to insert plugins at the correct position
    const nodes = React.Children.map(children, (child) => {
      // Check if this is the PluginSlot marker
      if (React.isValidElement(child) && child.type === PluginSlot) {
        return (
          <>
            {pluginSlice.map((ps) => {
              const PluginFC = ps.component as FC<typeof props>;
              return <PluginFC key={ps.info.slug_name} {...props} />;
            })}
            {pluginSlice.length > 0 && <div className="toolbar-divider" />}
          </>
        );
      }
      return child;
    });

    return <div className={className}>{nodes}</div>;
  }

  return (
    <>
      {pluginSlice.map((ps) => {
        const PluginFC = ps.component as FC<
          { className?: string } & typeof props
        >;
        return (
          <PluginFC key={ps.info.slug_name} className={className} {...props} />
        );
      })}
    </>
  );
};

export default Index;
