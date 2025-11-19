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

import React, {
  ForwardRefRenderFunction,
  forwardRef,
  useImperativeHandle,
  useEffect,
} from 'react';
import { Form, Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import isEmpty from 'lodash/isEmpty';
import classnames from 'classnames';

import { scrollElementIntoView } from '@/utils';
import type * as Type from '@/common/interface';

import type {
  JSONSchema,
  FormProps,
  FormRef,
  BaseUIOptions,
  FormKit,
  InputGroupOptions,
} from './types';
import {
  Legend,
  Select,
  Check,
  Switch,
  Timezone,
  Upload,
  Textarea,
  Input,
  Button as SfButton,
  InputGroup,
  TagSelector,
} from './components';

export * from './types';

/**
 * TODO:
 *  - [!] Standardised `Admin/Plugins/Config/index.tsx` method for generating dynamic form configurations.
 *  - Normalize and document `formData[key].hidden && 'd-none'`
 *  - Normalize and document `hiddenSubmit`
 *  - Improving field hints for `formData`
 *  - Optimise form data updates
 *    * Automatic field type conversion
 *    * Dynamic field generation
 */

/**
 * json schema form
 * @param schema json schema
 * @param uiSchema ui schema
 * @param formData form data
 * @param onChange change event
 * @param onSubmit submit event
 */
const SchemaForm: ForwardRefRenderFunction<FormRef, FormProps> = (
  {
    schema,
    uiSchema = {},
    refreshConfig,
    formData,
    onChange,
    onSubmit,
    hiddenSubmit = false,
  },
  ref,
) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'form',
  });
  const { required = [], properties = {} } = schema || {};
  // check required field
  const excludes = required.filter((key) => !properties[key]);
  if (excludes.length > 0) {
    console.error(t('not_found_props', { key: excludes.join(', ') }));
  }
  formData ||= {};
  const keys = Object.keys(properties);
  /**
   * Prevent components such as `select` from having default values,
   * which are not generated on `formData`
   */
  const setDefaultValueAsDomBehaviour = () => {
    keys.forEach((k) => {
      const fieldVal = formData![k]?.value;
      const metaProp = properties[k];
      const uiCtrl = uiSchema[k]?.['ui:widget'];
      if (!metaProp || !uiCtrl || fieldVal !== undefined) {
        return;
      }
      if (uiCtrl === 'select' && metaProp.enum?.[0] !== undefined) {
        formData![k] = {
          errorMsg: '',
          isInvalid: false,
          value: metaProp.enum?.[0],
        };
      }
    });
  };
  useEffect(() => {
    setDefaultValueAsDomBehaviour();
  }, [formData]);

  const formKitWithContext: FormKit = {
    refreshConfig() {
      if (typeof refreshConfig === 'function') {
        refreshConfig();
      }
    },
  };

  /**
   * Form validation
   * - Currently only dynamic forms are in use, the business form validation has been handed over to the server
   */
  const requiredValidator = () => {
    const errors: string[] = [];
    required.forEach((key) => {
      if (!formData![key] || !formData![key].value) {
        errors.push(key);
      }
    });
    return errors;
  };

  const syncValidator = () => {
    const errors: Array<{ key: string; msg: string }> = [];
    const promises: Array<{
      key: string;
      promise;
    }> = [];
    keys.forEach((key) => {
      const { validator } = uiSchema[key]?.['ui:options'] || {};
      if (validator instanceof Function) {
        const value = formData![key]?.value;
        promises.push({
          key,
          promise: validator(value, formData),
        });
      }
    });
    return Promise.allSettled(promises.map((item) => item.promise)).then(
      (results) => {
        results.forEach((result, index) => {
          const { key } = promises[index];
          if (result.status === 'rejected') {
            errors.push({
              key,
              msg: result.reason.message,
            });
          }

          if (result.status === 'fulfilled') {
            const msg = result.value;
            if (typeof msg === 'string') {
              errors.push({
                key,
                msg,
              });
            }
          }
        });
        return errors;
      },
    );
  };

  const validator = async (): Promise<boolean> => {
    const errors = requiredValidator();
    if (errors.length > 0) {
      formData = errors.reduce((acc, cur) => {
        acc[cur] = {
          ...formData![cur],
          isInvalid: true,
          errorMsg:
            uiSchema[cur]?.['ui:options']?.empty ||
            `${properties[cur]?.title} ${t('empty')}`,
        };
        return acc;
      }, formData || {});
      if (onChange instanceof Function) {
        onChange({ ...formData });
      }
      scrollElementIntoView(document.getElementById(errors[0]));
      return false;
    }
    const syncErrors = await syncValidator();
    if (syncErrors.length > 0) {
      formData = syncErrors.reduce((acc, cur) => {
        acc[cur.key] = {
          ...formData![cur.key],
          isInvalid: true,
          errorMsg: cur.msg || `${properties[cur.key].title} ${t('invalid')}`,
        };
        return acc;
      }, formData || {});
      if (onChange instanceof Function) {
        onChange({ ...formData });
      }
      scrollElementIntoView(document.getElementById(syncErrors[0].key));
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const isValid = await validator();
    if (!isValid) {
      return;
    }

    Object.keys(formData!).forEach((key) => {
      formData![key].isInvalid = false;
      formData![key].errorMsg = '';
    });
    if (onChange instanceof Function) {
      onChange(formData!);
    }
    if (onSubmit instanceof Function) {
      onSubmit(e);
    }
  };

  useImperativeHandle(ref, () => ({
    validator,
  }));
  if (!formData || !schema || isEmpty(schema.properties)) {
    return null;
  }

  return (
    <Form noValidate onSubmit={handleSubmit}>
      {keys.map((key) => {
        const {
          title,
          description,
          enum: enumValues = [],
          enumNames = [],
          max_length = 0,
          max,
          min = 0,
        } = properties[key];
        const { 'ui:widget': widget = 'input', 'ui:options': uiOpt } =
          uiSchema?.[key] || {};
        formData ||= {};
        const fieldState = formData[key];
        if (uiOpt?.class_name) {
          uiOpt.className = uiOpt.class_name;
        }

        const uiSimplify = widget === 'legend' || uiOpt?.simplify;
        let groupClassName: BaseUIOptions['field_class_name'] = uiOpt?.simplify
          ? 'mb-2'
          : 'mb-3';
        if (widget === 'legend') {
          groupClassName = 'mb-0';
        }
        if (uiOpt?.field_class_name) {
          groupClassName = uiOpt.field_class_name;
        }

        const readOnly = uiOpt?.readOnly || false;

        return (
          <Form.Group
            key={`${title}-${key}`}
            controlId={key}
            className={classnames(
              groupClassName,
              formData[key].hidden ? 'd-none' : null,
            )}>
            {/* Uniform processing `label` */}
            {title && !uiSimplify ? <Form.Label>{title}</Form.Label> : null}
            {/* Handling of individual specific controls */}
            {widget === 'legend' ? (
              <Legend title={title} className={String(uiOpt?.className)} />
            ) : null}
            {widget === 'select' ? (
              <Select
                desc={description}
                fieldName={key}
                onChange={onChange}
                enumValues={enumValues}
                enumNames={enumNames}
                formData={formData}
                readOnly={readOnly}
              />
            ) : null}
            {widget === 'radio' || widget === 'checkbox' ? (
              <Check
                type={widget}
                fieldName={key}
                onChange={onChange}
                enumValues={enumValues}
                enumNames={enumNames}
                formData={formData}
                readOnly={readOnly}
              />
            ) : null}
            {widget === 'switch' ? (
              <Switch
                label={uiOpt && 'label' in uiOpt ? uiOpt.label : ''}
                fieldName={key}
                onChange={onChange}
                formData={formData}
                readOnly={readOnly}
              />
            ) : null}
            {widget === 'timezone' ? (
              <Timezone
                fieldName={key}
                onChange={onChange}
                formData={formData}
                readOnly={readOnly}
              />
            ) : null}
            {widget === 'upload' ? (
              <Upload
                type={
                  uiOpt && 'imageType' in uiOpt ? uiOpt.imageType : undefined
                }
                acceptType={
                  uiOpt && 'acceptType' in uiOpt ? uiOpt.acceptType : ''
                }
                fieldName={key}
                onChange={onChange}
                formData={formData}
                readOnly={readOnly}
                imgClassNames={
                  uiOpt && 'className' in uiOpt ? uiOpt.className : ''
                }
              />
            ) : null}
            {widget === 'textarea' ? (
              <Textarea
                placeholder={
                  uiOpt && 'placeholder' in uiOpt ? uiOpt.placeholder : ''
                }
                rows={uiOpt && 'rows' in uiOpt ? uiOpt.rows : 3}
                className={uiOpt && 'className' in uiOpt ? uiOpt.className : ''}
                fieldName={key}
                onChange={onChange}
                formData={formData}
                readOnly={readOnly}
              />
            ) : null}
            {widget === 'input' ? (
              <Input
                type={uiOpt && 'inputType' in uiOpt ? uiOpt.inputType : 'text'}
                inputMode={
                  uiOpt && 'inputMode' in uiOpt ? uiOpt.inputMode : 'text'
                }
                placeholder={
                  uiOpt && 'placeholder' in uiOpt ? uiOpt.placeholder : ''
                }
                min={min}
                max={max}
                fieldName={key}
                onChange={onChange}
                formData={formData}
                readOnly={readOnly}
              />
            ) : null}
            {widget === 'button' ? (
              <SfButton
                fieldName={key}
                text={uiOpt && 'text' in uiOpt ? uiOpt.text : ''}
                action={uiOpt && 'action' in uiOpt ? uiOpt.action : undefined}
                formKit={formKitWithContext}
                readOnly={readOnly}
                variant={
                  uiOpt && 'variant' in uiOpt ? uiOpt.variant : undefined
                }
                size={uiOpt && 'size' in uiOpt ? uiOpt.size : undefined}
                title={uiOpt && 'title' in uiOpt ? uiOpt?.title : ''}
              />
            ) : null}
            {widget === 'input_group' ? (
              <InputGroup
                formKitWithContext={formKitWithContext}
                uiOpt={uiOpt as InputGroupOptions}
                prefixText={
                  (uiOpt && 'prefixText' in uiOpt && uiOpt.prefixText) || ''
                }
                suffixText={
                  (uiOpt && 'suffixText' in uiOpt && uiOpt.suffixText) || ''
                }>
                <Input
                  type={
                    uiOpt && 'inputType' in uiOpt ? uiOpt.inputType : 'text'
                  }
                  inputMode={
                    uiOpt && 'inputMode' in uiOpt ? uiOpt.inputMode : 'text'
                  }
                  placeholder={
                    uiOpt && 'placeholder' in uiOpt ? uiOpt.placeholder : ''
                  }
                  min={min}
                  max={max}
                  fieldName={key}
                  onChange={onChange}
                  formData={formData}
                  readOnly={readOnly}
                />
              </InputGroup>
            ) : null}
            {widget === 'tag_selector' ? (
              <TagSelector
                maxTagLength={max_length}
                fieldName={key}
                onChange={onChange}
                formData={formData}
                description={description}
              />
            ) : null}
            {/* Unified handling of `Feedback` and `Text` */}
            <Form.Control.Feedback type="invalid">
              {fieldState?.errorMsg}
            </Form.Control.Feedback>
            {description && widget !== 'tag_selector' ? (
              <Form.Text dangerouslySetInnerHTML={{ __html: description }} />
            ) : null}
          </Form.Group>
        );
      })}
      {!hiddenSubmit && (
        <Button variant="primary" type="submit">
          {t('btn_submit')}
        </Button>
      )}
    </Form>
  );
};

export const initFormData = (schema: JSONSchema): Type.FormDataType => {
  const formData: Type.FormDataType = {};
  const props: JSONSchema['properties'] = schema?.properties || {};
  Object.keys(props).forEach((key) => {
    const prop = props[key];
    let defaultVal: any = '';
    if (Array.isArray(prop.default) && prop.enum && prop.enum.length > 0) {
      // for checkbox default values
      defaultVal = prop.enum;
    } else {
      defaultVal = prop?.default;
    }
    formData[key] = {
      value: defaultVal,
      isInvalid: false,
      errorMsg: '',
    };
  });
  return formData;
};

export const mergeFormData = (
  target: Type.FormDataType | null,
  origin: Type.FormDataType | null,
) => {
  if (!target) {
    return origin;
  }
  if (!origin) {
    return target;
  }
  Object.keys(target).forEach((k) => {
    const oi = origin[k];
    if (oi && oi.value !== undefined) {
      target[k] = {
        value: oi.value,
        isInvalid: false,
        errorMsg: '',
      };
    }
  });
  return target;
};

export default forwardRef(SchemaForm);
