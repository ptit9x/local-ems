import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import './styles.css';
import { App } from './App';

createRoot(document.getElementById('root')!).render(
  <StrictMode><QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { refetchOnWindowFocus: false, retry: 1 } } })}><App /></QueryClientProvider></StrictMode>,
);
