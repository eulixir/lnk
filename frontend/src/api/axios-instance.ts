import axios, { AxiosRequestConfig, AxiosResponse } from "axios";

// Create axios instance with base URL from environment variable
// Note: In Next.js, client-side code can only access NEXT_PUBLIC_* env variables
// Set NEXT_PUBLIC_API_URL for client-side usage
// For server-side, you can use API_URL, but you'll need to set NEXT_PUBLIC_API_URL as well
const baseURL = process.env.NEXT_PUBLIC_API_URL || process.env.API_URL || "http://localhost:8080";

const axiosInstance = axios.create({
  baseURL,
});

// Custom instance function for orval
// Returns the full AxiosResponse to match what orval expects
export const customInstance = <T>(
  config: AxiosRequestConfig,
  options?: AxiosRequestConfig,
): Promise<AxiosResponse<T>> => {
  const source = axios.CancelToken.source();
  const promise = axiosInstance({
    ...config,
    ...options,
    cancelToken: source.token,
  });

  // @ts-ignore
  promise.cancel = () => {
    source.cancel("Query was cancelled");
  };

  return promise;
};

export default axiosInstance;

