import axios from 'axios';

const client = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

client.interceptors.response.use(
  (res) => res,
  (err) => {
    if (!err.response) {
      console.warn('[API] Network error or timeout');
    }
    return Promise.reject(err);
  }
);

export default client;
