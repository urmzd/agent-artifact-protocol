// Data Hooks

import { useState, useEffect } from 'react';
import { AbortController } from 'abort-controller';

// useApi
export function useApi<T>(url: string): {
  data: T | null;
  error: Error | null;
  isLoading: boolean;
} {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  useEffect(() => {
    const controller = new AbortController();
    const signal = controller.signal;

    setIsLoading(true);
    fetch(url, { signal })
      .then((response) => {
        if (!response.ok) {
          throw new Error(response.statusText);
        }
        return response.json() as Promise<T>;
      })
      .then((data) => {
        setData(data);
        setError(null);
      })
      .catch((error) => {
        if (error.name === 'AbortError') {
          return;
        }
        setError(error);
        setData(null);
      })
      .finally(() => {
        setIsLoading(false);
      });

    return () => {
      controller.abort();
    };
  }, [url]);

  return { data, error, isLoading };
}

// usePagination
export function usePagination<T>(data: T[], itemsPerPage: number): {
  paginatedData: T[];
  currentPage: number;
  totalPages: number;
  handlePageChange: (page: number) => void;
} {
  const [currentPage, setCurrentPage] = useState(1);

  const start = (currentPage - 1) * itemsPerPage;
  const end = start + itemsPerPage;
  const paginatedData = data.slice(start, end);
  const totalPages = Math.ceil(data.length / itemsPerPage);

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  return { paginatedData, currentPage, totalPages, handlePageChange };
}

// useInfiniteScroll
export function useInfiniteScroll<T>(data: T[], fetchMore: () => void): {
  isInfiniteLoading: boolean;
  handleInfiniteScroll: () => void;
} {
  const [isInfiniteLoading, setIsInfiniteLoading] = useState(false);

  const handleInfiniteScroll = () => {
    fetchMore();
    setIsInfiniteLoading(true);
  };

  useEffect(() => {
    const handleScroll = () => {
      if (window.innerHeight + window.scrollY >= document.body.offsetHeight) {
        handleInfiniteScroll();
      }
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [data]);

  useEffect(() => {
    if (isInfiniteLoading) {
      const timer = setTimeout(() => {
        setIsInfiniteLoading(false);
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [isInfiniteLoading]);

  return { isInfiniteLoading, handleInfiniteScroll };
}

// useDebounce
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}

// useLocalStorage
export function useLocalStorage<T>(key: string, initialValue: T): {
  value: T;
  setValue: (newValue: T) => void;
} {
  const [value, setValue] = useState<T>(() => {
    const storedValue = localStorage.getItem(key);
    return storedValue !== null ? JSON.parse(storedValue) : initialValue;
  });

  useEffect(() => {
    localStorage.setItem(key, JSON.stringify(value));
  }, [value, key]);

  return { value, setValue };
}

// UI Hooks

// useMediaQuery
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(false);

  useEffect(() => {
    const mediaQuery = window.matchMedia(query);
    setMatches(mediaQuery.matches);
    const listener = () => {
      setMatches(mediaQuery.matches);
    };
    mediaQuery.addEventListener('change', listener);
    return () => mediaQuery.removeEventListener('change', listener);
  }, [query]);

  return matches;
}

// useClickOutside
export function useClickOutside(ref: React.RefObject<HTMLElement>, handler: () => void): void {
  useEffect(() => {
    const listener = (event: MouseEvent) => {
      if (!ref.current || !ref.current.contains(event.target as Node)) {
        handler();
      }
    };
    document.addEventListener('mousedown', listener);
    return () => document.removeEventListener('mousedown', listener);
  }, [ref, handler]);
}

// useKeyboardShortcut
export function useKeyboardShortcut(keyCode: number, handler: () => void): void {
  useEffect(() => {
    const listener = (event: KeyboardEvent) => {
      if (event.keyCode === keyCode) {
        handler();
      }
    };
    document.addEventListener('keydown', listener);
    return () => document.removeEventListener('keydown', listener);
  }, [keyCode, handler]);
}

// useTheme
export function useTheme(): {
  theme: 'light' | 'dark';
  toggleTheme: () => void;
} {
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    const storedTheme = localStorage.getItem('theme');
    return storedTheme === 'dark' ? 'dark' : 'light';
  });

  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : 'light');
  };

  useEffect(() => {
    localStorage.setItem('theme', theme);
    document.body.className = theme;
  }, [theme]);

  return { theme, toggleTheme };
}

// useToast
export function useToast(): {
  toast: (message: string) => void;
  message: string | null;
} {
  const [message, setMessage] = useState<string | null>(null);

  const toast = (message: string) => {
    setMessage(message);
    const timer = setTimeout(() => {
      setMessage(null);
    }, 2000);
    return () => clearTimeout(timer);
  };

  return { toast, message };
}

// Form Hooks

// useForm
export function useForm<T>(initialValue: T): {
  value: T;
  setValue: (newValue: T) => void;
  handleChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
} {
  const [value, setValue] = useState<T>(initialValue);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const updatedValue = { ...value };
    updatedValue[event.target.name as keyof T] = event.target.value;
    setValue(updatedValue);
  };

  return { value, setValue, handleChange };
}

// useFieldValidation
export function useFieldValidation<T>(field: keyof T, value: T, validate: (value: T[keyof T]) => string | null): string | null {
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const error = validate(value[field]);
    setError(error);
  }, [field, value, validate]);

  return error;
}

// useFormSubmit
export function useFormSubmit<T>(handleSubmit: (value: T) => void): (event: React.FormEvent<HTMLFormElement>) => void {
  const submitHandler = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    handleSubmit(event.currentTarget.elements);
  };

  return submitHandler;
}