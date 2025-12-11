/**
 * App Component
 * Root application component with routing
 */

import { useEffect } from 'react';
import { ErrorBoundary } from './components/ErrorBoundary';
import { ToastContainer } from './components/ToastContainer';
import { ThemeProvider } from './components/ThemeProvider';
import { AppRouter } from './routes';
import { authStore } from './stores/authStore';
import './App.css';

function App() {
  // Initialize auth on app startup
  useEffect(() => {
    authStore.initialize();
  }, []);

  return (
    <ThemeProvider>
      <ErrorBoundary>
        <AppRouter />
        <ToastContainer />
      </ErrorBoundary>
    </ThemeProvider>
  );
}

export default App;

