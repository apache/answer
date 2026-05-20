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

import { FC, useContext, useEffect } from 'react';
import { Dropdown, Button } from 'react-bootstrap';

import { EditorContext } from './EditorContext';
import { Editor } from './types';

interface IProps {
  keyMap?: string[];
  onClick?: (editor: Editor) => void;
  tip?: string;
  className?: string;
  as?: any;
  children?;
  label?: string;
  disable?: boolean;
  isShow?: boolean;
  onBlur?: (editor: Editor) => void;
}
const ToolItem: FC<IProps> = (props) => {
  const editor = useContext(EditorContext);

  const {
    label,
    tip,
    disable = false,
    isShow,
    keyMap,
    onClick,
    className,
    as,
    children,
    onBlur,
  } = props;

  useEffect(() => {
    if (!editor) {
      return;
    }
    if (!keyMap) {
      return;
    }

    keyMap.forEach((key) => {
      editor?.addKeyMap({
        [key]: () => {
          onClick?.(editor);
          return true;
        },
      });
    });
  }, [editor]);

  const btnRender = () => (
    <Button
      variant="link"
      title={tip}
      className={`p-0 b-0 btn-no-border toolbar text-body ${
        disable ? 'disabled' : ''
      }`}
      disabled={disable}
      tabIndex={-1}
      onClick={(e) => {
        e.preventDefault();
        if (editor) {
          onClick?.(editor);
        }
      }}
      onBlur={(e) => {
        e.preventDefault();
        if (editor) {
          onBlur?.(editor);
        }
      }}>
      <i className={`bi bi-${label}`} />
    </Button>
  );

  if (!editor) {
    return null;
  }
  return (
    <div className={`toolbar-item-wrap ${className || ''}`}>
      {as === 'dropdown' ? (
        <Dropdown className="h-100 w-100" show={isShow}>
          <Dropdown.Toggle as="div" className="h-100">
            {btnRender()}
          </Dropdown.Toggle>
          {children}
        </Dropdown>
      ) : (
        <>
          {btnRender()}
          {children}
        </>
      )}
    </div>
  );
};

export default ToolItem;
