// ui hooks
import { useState, useEffect, useRef, useCallback } from 'react';

export function useToast() {
  const [toasts, setToasts] = useState<{ message: string; id: string; timeoutId: number | null }[]>([]);
  const [paused, setPaused] = useState(false);
  const toastContainerRef = useRef<HTMLDivElement | null>(null);
  const toastId = useRef(0);

  const showToast = useCallback((message: string) => {
    const newToast = { message, id: `${toastId.current}`, timeoutId: null };
    setToasts((prevToasts) => {
      const updatedToasts = [...prevToasts, newToast];
      if (updatedToasts.length > 5) {
        updatedToasts.shift();
      }
      return updatedToasts;
    });
    toastId.current++;
  }, []);

  const handleMouseEnter = useCallback(() => {
    setPaused(true);
    toasts.forEach((toast) => {
      if (toast.timeoutId) {
        globalThis.clearTimeout(toast.timeoutId);
      }
    });
  }, [toasts]);

  const handleMouseLeave = useCallback(() => {
    setPaused(false);
    toasts.forEach((toast) => {
      if (!toast.timeoutId) {
        const timeoutId = globalThis.setTimeout(() => {
          setToasts((prevToasts) => prevToasts.filter((t) => t.id !== toast.id));
        }, 5000);
        setToasts((prevToasts) =>
          prevToasts.map((t) => (t.id === toast.id ? { ...t, timeoutId } : t))
        );
      }
    });
  }, [toasts]);

  useEffect(() => {
    if (toastContainerRef.current) {
      toastContainerRef.current.addEventListener('mouseenter', handleMouseEnter);
      toastContainerRef.current.addEventListener('mouseleave', handleMouseLeave);
      return () => {
        toastContainerRef.current.removeEventListener('mouseenter', handleMouseEnter);
        toastContainerRef.current.removeEventListener('mouseleave', handleMouseLeave);
      };
    }
  }, [handleMouseEnter, handleMouseLeave]);

  useEffect(() => {
    if (!paused) {
      toasts.forEach((toast) => {
        if (!toast.timeoutId) {
          const timeoutId = globalThis.setTimeout(() => {
            setToasts((prevToasts) => prevToasts.filter((t) => t.id !== toast.id));
          }, 5000);
          setToasts((prevToasts) =>
            prevToasts.map((t) => (t.id === toast.id ? { ...t, timeoutId } : t))
          );
        }
      });
    }
  }, [toasts, paused]);

  return { toasts, showToast, toastContainerRef };
}