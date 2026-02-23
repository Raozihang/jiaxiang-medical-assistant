export type ApiResponse<T> = {
  code: number;
  message: string;
  data: T;
  request_id: string;
  timestamp: string;
};

export type Paginated<T> = {
  items: T[];
  page: number;
  page_size: number;
  total: number;
};
