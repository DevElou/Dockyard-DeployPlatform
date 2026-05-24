import { toast } from "sonner";
import { ApiError } from "@/lib/types/api";
import type { UseFormReturn } from "react-hook-form";

export function getErrorMessage(error: unknown): string {
  if (error instanceof ApiError) return error.message;
  if (error instanceof Error) return error.message;
  return "Une erreur inattendue s'est produite.";
}

type FieldMap = Record<string, string>;

export function handleMutationError(
  error: unknown,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form?: UseFormReturn<any>,
  fieldMap: FieldMap = {},
): void {
  if (error instanceof ApiError) {
    if ((error.status === 400 || error.status === 409) && form) {
      const field = fieldMap[error.code];
      if (field) {
        form.setError(field, { message: error.message });
        return;
      }
      const msgLower = error.message.toLowerCase();
      for (const [key, fieldName] of Object.entries(fieldMap)) {
        if (msgLower.includes(key)) {
          form.setError(fieldName, { message: error.message });
          return;
        }
      }
    }
    toast.error(error.message);
    return;
  }
  toast.error("Une erreur inattendue s'est produite.");
}
