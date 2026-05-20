# Schema Form

## Introduction

A React component capable of building HTML forms out of a [JSON schema](https://json-schema.org/understanding-json-schema/index.html).

## Usage

```tsx
import React, { useState } from 'react';

import {
  SchemaForm,
  initFormData,
  JSONSchema,
  UISchema,
  FormKit,
} from '@/components';

const schema: JSONSchema = {
  title: 'General',
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    age: {
      type: 'number',
      title: 'Age',
    },
    sex: {
      type: 'string',
      title: 'sex',
      enum: [1, 2],
      enumNames: ['male', 'female'],
    },
  },
};

const uiSchema: UISchema = {
  name: {
    'ui:widget': 'input',
  },
  age: {
    'ui:widget': 'input',
    'ui:options': {
      type: 'number',
    },
  },
  sex: {
    'ui:widget': 'radio',
  },
};

const Form = () => {
  const [formData, setFormData] = useState(initFormData(schema));

  const formRef = useRef<{
    validator: () => Promise<boolean>;
  }>(null);

  const refreshConfig: FormKit['refreshConfig'] = async () => {
    // refreshFormConfig();
  };

  const handleChange = (data) => {
    setFormData(data);
  };

  return (
    <SchemaForm
      ref={formRef}
      schema={schema}
      uiSchema={uiSchema}
      formData={formData}
      onChange={handleChange}
      refreshConfig={refreshConfig}
    />
  );
};

export default Form;
```

---

## Form Props

```ts
interface FormProps {
  // Describe the form structure with schema
  schema: JSONSchema | null;
  // Describe the properties of the field
  uiSchema?: UISchema;
  // Describe form data
  formData: Type.FormDataType | null;
  // Callback function when form data changes
  onChange?: (data: Type.FormDataType) => void;
  // Handler for when a form fires a `submit` event
  onSubmit?: (e: React.FormEvent) => void;
  /**
   * Callback method for updating form configuration
   * information (schema/uiSchema) in UIAction
   */
  refreshConfig?: FormKit['refreshConfig'];
}
```

## Form Ref

```ts
export interface FormRef {
  validator: () => Promise<boolean>;
}
```

When you need to validate a form and get the result outside the form, you can create a `FormRef` with `useRef` and pass it to the form using the `ref` property.

This allows you to validate the form and get the result outside the form using `formRef.current.validator()`.

---

## Types Definition

### JSONSchema

```ts
export interface JSONSchema {
  title: string;
  description?: string;
  required?: string[];
  properties: {
    [key: string]: {
      type: 'string' | 'boolean' | 'number';
      title: string;
      label?: string;
      description?: string;
      enum?: Array<string | boolean | number>;
      enumNames?: string[];
      default?: string | boolean | number;
    };
  };
}
```

### UISchema

```ts
export interface UISchema {
  [key: string]: {
    'ui:widget'?: UIWidget;
    'ui:options'?: UIOptions;
  };
}
```

### UIWidget

```ts
export type UIWidget =
  | 'textarea'
  | 'input'
  | 'checkbox'
  | 'radio'
  | 'select'
  | 'upload'
  | 'timezone'
  | 'switch'
  | 'legend'
  | 'button';
```

---

### UIOptions

```ts
export type UIOptions =
  | InputOptions
  | SelectOptions
  | UploadOptions
  | SwitchOptions
  | TimezoneOptions
  | CheckboxOptions
  | RadioOptions
  | TextareaOptions
  | ButtonOptions;
```

#### BaseUIOptions

```ts
export interface BaseUIOptions {
  empty?: string;
  // Will be appended to the className of the form component itself
  className?: classnames.Argument;
  class_name?: classnames.Argument;
  // The className that will be attached to a form field container
  field_class_name?: classnames.Argument;
  // Make a form component render into simplified mode
  readOnly?: boolean;
  simplify?: boolean;
  validator?: (
    value,
    formData?,
  ) => Promise<string | true | void> | true | string;
}
```

#### InputOptions

```ts
export interface InputOptions extends BaseUIOptions {
  placeholder?: string;
  inputType?:
    | 'color'
    | 'date'
    | 'datetime-local'
    | 'email'
    | 'month'
    | 'number'
    | 'password'
    | 'range'
    | 'search'
    | 'tel'
    | 'text'
    | 'time'
    | 'url'
    | 'week';
}
```

#### SelectOptions

```ts
export interface SelectOptions extends UIOptions {}
```

#### UploadOptions

```ts
export interface UploadOptions extends BaseUIOptions {
  acceptType?: string;
  imageType?: Type.UploadType;
}
```

#### SwitchOptions

```ts
export interface SwitchOptions extends BaseUIOptions {
  label?: string;
}
```

#### TimezoneOptions

```ts
export interface TimezoneOptions extends UIOptions {
  placeholder?: string;
}
```

#### CheckboxOptions

```ts
export interface CheckboxOptions extends UIOptions {}
```

#### RadioOptions

```ts
export interface RadioOptions extends UIOptions {}
```

#### TextareaOptions

```ts
export interface TextareaOptions extends UIOptions {
  placeholder?: string;
  rows?: number;
}
```

#### ButtonOptions

```ts
export interface ButtonOptions extends BaseUIOptions {
  text: string;
  icon?: string;
  action?: UIAction;
  variant?: ButtonProps['variant'];
  size?: ButtonProps['size'];
}
```

#### UIAction

```ts
export interface UIAction {
  url: string;
  method?: 'get' | 'post' | 'put' | 'delete';
  loading?: {
    text: string;
    state?: 'none' | 'pending' | 'completed';
  };
  on_complete?: {
    toast_return_message?: boolean;
    refresh_form_config?: boolean;
  };
}
```

#### FormKit

```ts
export interface FormKit {
  refreshConfig(): void;
}
```

---

### FormData

```ts
export interface FormValue<T = any> {
  value: T;
  isInvalid: boolean;
  errorMsg: string;
  [prop: string]: any;
}

export interface FormDataType {
  [prop: string]: FormValue;
}
```

---

## Backend API

For backend generating modal form you can return json like this.

### Response

```json
{
  "name": "string",
  "slug_name": "string",
  "description": "string",
  "version": "string",
  "config_fields": [
    {
      "name": "string",
      "type": "textarea",
      "title": "string",
      "description": "string",
      "required": true,
      "value": "string",
      "ui_options": {
        "placeholder": "placeholder",
        "rows": 4
      },
      "options": [
        {
          "value": "string",
          "label": "string"
        }
      ]
    }
  ]
}
```

## reference

- [json schema](https://json-schema.org/understanding-json-schema/index.html)
- [react-jsonschema-form](https://github.com/rjsf-team/react-jsonschema-form)
- [vue-json-schema-form](https://github.com/lljj-x/vue-json-schema-form/)
