// form hooks
import { useState, useEffect, useRef, useCallback } from 'react';

export function useForm<T>(initialState: T) {
  const [formState, setFormState] = useState<T>(initialState);

  const getNestedValue = useCallback((obj: any, path: string) => {
    const keys = path.split('.');
    let current = obj;
    for (let i = 0; i < keys.length; i++) {
      if (!current) return undefined;
      const key = keys[i];
      const match = key.match(/^([^\[\]]+)(\[(\d+)\])?$/);
      if (match) {
        const propName = match[1];
        const arrayIndex = match[3] ? parseInt(match[3], 10) : undefined;
        if (arrayIndex !== undefined) {
          if (!Array.isArray(current[propName])) {
            current[propName] = [];
          }
          if (i === keys.length - 1) {
            return current[propName][arrayIndex];
          } else {
            current = current[propName][arrayIndex];
          }
        } else {
          if (i === keys.length - 1) {
            return current[propName];
          } else {
            current = current[propName];
          }
        }
      }
    }
  }, []);

  const setNestedValue = useCallback((obj: any, path: string, value: any) => {
    const keys = path.split('.');
    let current = obj;
    for (let i = 0; i < keys.length; i++) {
      if (!current) current = {};
      const key = keys[i];
      const match = key.match(/^([^\[\]]+)(\[(\d+)\])?$/);
      if (match) {
        const propName = match[1];
        const arrayIndex = match[3] ? parseInt(match[3], 10) : undefined;
        if (arrayIndex !== undefined) {
          if (!Array.isArray(current[propName])) {
            current[propName] = [];
          }
          if (i === keys.length - 1) {
            if (arrayIndex >= current[propName].length) {
              current[propName].push(...new Array(arrayIndex - current[propName].length + 1).fill(null));
            }
            current[propName][arrayIndex] = value;
          } else {
            if (arrayIndex >= current[propName].length) {
              current[propName].push(...new Array(arrayIndex - current[propName].length + 1).fill(null));
            }
            current = current[propName][arrayIndex];
          }
        } else {
          if (i === keys.length - 1) {
            current[propName] = value;
          } else {
            if (!current[propName]) current[propName] = {};
            current = current[propName];
          }
        }
      }
    }
  }, []);

  const handleFormChange = useCallback((event: React.ChangeEvent<{ name: string; value: any }>) => {
    const { name, value } = event.target;
    setFormState((prevFormState) => {
      const newState = { ...prevFormState };
      setNestedValue(newState, name, value);
      return newState;
    });
  }, [setNestedValue]);

  return { formState, handleFormChange };
}

export function useFormArray<T>(initialState: T[], setName: string) {
  const [arrayState, setArrayState] = useState<T[]>(initialState);

  const handleArrayChange = useCallback((index: number, value: T) => {
    setArrayState((prevArrayState) => {
      const newArrayState = [...prevArrayState];
      newArrayState[index] = value;
      return newArrayState;
    });
  }, []);

  const handleArrayAdd = useCallback((value: T) => {
    setArrayState((prevArrayState) => {
      return [...prevArrayState, value];
    });
  }, []);

  const handleArrayRemove = useCallback((index: number) => {
    setArrayState((prevArrayState) => {
      return prevArrayState.filter((item, i) => i !== index);
    });
  }, []);

  return { arrayState, handleArrayChange, handleArrayAdd, handleArrayRemove };
}

// existing hooks...