// data hooks
import { useState, useEffect, useRef, useCallback } from 'react';

export function useApi<T>(url: string, params?: { [key: string]: any }) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const abortController = useRef<AbortController | null>(null);

  const fetchApi = useCallback(async () => {
    if (abortController.current) {
      abortController.current.abort();
    }
    abortController.current = new AbortController();
    setLoading(true);
    try {
      const response = await fetch(url, {
        signal: abortController.current.signal,
        ...params,
      });
      if (!response.ok) {
        throw new Error(response.statusText);
      }
      const data = await response.json();
      setData(data);
    } catch (error) {
      setError(error);
    } finally {
      setLoading(false);
    }
  }, [url, params]);

  useEffect(() => {
    fetchApi();
    return () => {
      if (abortController.current) {
        abortController.current.abort();
      }
    };
  }, [fetchApi]);

  return { data, loading, error };
}

export function usePagination<T>(data: T[], itemsPerPage: number) {
  const [currentPage, setCurrentPage] = useState(1);
  const start = (currentPage - 1) * itemsPerPage;
  const end = start + itemsPerPage;
  const paginatedData = data.slice(start, end);

  return { paginatedData, currentPage, setCurrentPage };
}

export function useInfiniteScroll<T>(data: T[], fetchMore: () => void) {
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const scrollElement = useRef<HTMLDivElement | null>(null);

  const handleScroll = useCallback(() => {
    if (scrollElement.current) {
      const { scrollTop, offsetHeight, scrollHeight } = scrollElement.current;
      if (scrollTop + offsetHeight >= scrollHeight * 0.9 && hasMore && !loading) {
        fetchMore();
      }
    }
  }, [fetchMore, hasMore, loading]);

  useEffect(() => {
    if (scrollElement.current) {
      scrollElement.current.addEventListener('scroll', handleScroll);
    }
    return () => {
      if (scrollElement.current) {
        scrollElement.current.removeEventListener('scroll', handleScroll);
      }
    };
  }, [handleScroll]);

  return { scrollElement, setLoading, setHasMore };
}

export function useDebounce<T>(value: T, delay: number) {
  const [debouncedValue, setDebouncedValue] = useState(value);
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);
    return () => {
      clearTimeout(timeoutId);
    };
  }, [value, delay]);

  return debouncedValue;
}

export function useLocalStorage<T>(key: string, initialValue: T) {
  const [value, setValue] = useState<T>(() => {
    const storedValue = localStorage.getItem(key);
    return storedValue ? JSON.parse(storedValue) : initialValue;
  });

  useEffect(() => {
    localStorage.setItem(key, JSON.stringify(value));
  }, [value, key]);

  return { value, setValue };
}

// ui hooks
export function useMediaQuery(query: string) {
  const [matches, setMatches] = useState(false);
  const mediaQueryList = useRef<MediaQueryList | null>(null);

  const handleMediaQueryChange = useCallback((event: MediaQueryListEvent) => {
    setMatches(event.matches);
  }, []);

  useEffect(() => {
    mediaQueryList.current = window.matchMedia(query);
    setMatches(mediaQueryList.current.matches);
    mediaQueryList.current.addListener(handleMediaQueryChange);
    return () => {
      if (mediaQueryList.current) {
        mediaQueryList.current.removeListener(handleMediaQueryChange);
      }
    };
  }, [query, handleMediaQueryChange]);

  return matches;
}

export function useClickOutside(ref: React.RefObject<HTMLElement>) {
  const [isOutside, setIsOutside] = useState(false);

  const handleDocumentClick = useCallback((event: MouseEvent) => {
    if (!ref.current?.contains(event.target as Node)) {
      setIsOutside(true);
    } else {
      setIsOutside(false);
    }
  }, [ref]);

  useEffect(() => {
    document.addEventListener('click', handleDocumentClick);
    return () => {
      document.removeEventListener('click', handleDocumentClick);
    };
  }, [handleDocumentClick]);

  return isOutside;
}

export function useKeyboardShortcut(shortcut: string, callback: () => void) {
  const handleKeydown = useCallback((event: KeyboardEvent) => {
    if (event.key === shortcut) {
      callback();
    }
  }, [shortcut, callback]);

  useEffect(() => {
    document.addEventListener('keydown', handleKeydown);
    return () => {
      document.removeEventListener('keydown', handleKeydown);
    };
  }, [handleKeydown]);
}

export function useTheme() {
  const [theme, setTheme] = useState<'light' | 'dark'>('light');

  useEffect(() => {
    document.body.classList.remove('light', 'dark');
    document.body.classList.add(theme);
  }, [theme]);

  return { theme, setTheme };
}

export function useToast() {
  const [toast, setToast] = useState<{ message: string; duration: number } | null>(null);

  const showToast = useCallback((message: string, duration: number) => {
    setToast({ message, duration });
    setTimeout(() => {
      setToast(null);
    }, duration);
  }, []);

  return { toast, showToast };
}

// form hooks
export function useForm<T>(initialState: T) {
  const [formState, setFormState] = useState<T>(initialState);

  const handleFormChange = useCallback((event: React.ChangeEvent<{ name: string; value: any }>) => {
    setFormState((prevFormState) => ({ ...prevFormState, [event.target.name]: event.target.value }));
  }, []);

  return { formState, handleFormChange };
}

export function useFieldValidation<T>(field: string, validationFn: (value: T) => boolean) {
  const [isValid, setIsValid] = useState(true);
  const [error, setError] = useState('');

  const validateField = useCallback((value: T) => {
    const isValidField = validationFn(value);
    setIsValid(isValidField);
    if (!isValidField) {
      setError('Invalid field value');
    } else {
      setError('');
    }
  }, [validationFn]);

  return { isValid, error, validateField };
}

export function useFormSubmit<T>(formState: T, submitFn: (formState: T) => void) {
  const handleSubmit = useCallback((event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    submitFn(formState);
  }, [formState, submitFn]);

  return handleSubmit;
}