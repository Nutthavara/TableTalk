export type ApiErrorBody = {
  error: {
    code: string;
    message: string;
    details?: Record<string, string>;
  };
};

export type Me = {
  id: string;
  username: string;
  display_name: string;
  email?: string;
};
