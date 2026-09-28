// Reading typed values out of FormData in Server Actions. Values are passed to the API as-is; the
// API validates them. Local checks only catch values that cannot be sent as the right JSON type.

export function formString(fd: FormData, name: string): string {
  const v = fd.get(name);
  return typeof v === "string" ? v.trim() : "";
}

/** Parses a whole number, recording a field error when the value is not one. */
export function formInt(fd: FormData, name: string, errors: Record<string, string>): number {
  const raw = formString(fd, name);
  if (!/^-?\d+$/.test(raw)) {
    errors[name] = "Please enter a whole number.";
    return 0;
  }
  return Number(raw);
}

/** Non-empty uploaded files for a file input. */
export function formFiles(fd: FormData, name: string): File[] {
  return fd.getAll(name).filter((v): v is File => v instanceof File && v.size > 0);
}
