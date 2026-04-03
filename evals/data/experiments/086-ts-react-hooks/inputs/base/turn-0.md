Create custom React hooks in TypeScript for a dashboard application.

Include:
- Data hooks: useApi<T> (fetch with loading/error), usePagination, useInfiniteScroll, useDebounce, useLocalStorage<T>
- UI hooks: useMediaQuery, useClickOutside, useKeyboardShortcut, useTheme, useToast
- Form hooks: useForm<T> (generic form state), useFieldValidation, useFormSubmit
- Full TypeScript generics, proper cleanup in useEffect, AbortController for fetch cancellation

Use section IDs: data-hooks, ui-hooks, form-hooks

Use AAP section markers to delineate each major code block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.