export const FLASH_COOKIE = "tt_flash";

export type Flash = {
  message: string;
  kind?: "success" | "error";
  action?: { label: string; href: string };
};
