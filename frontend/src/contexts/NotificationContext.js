import { createContext, useContext, useState, useEffect, useCallback } from 'react';

const NotificationContext = createContext();

export function NotificationProvider({ children }) {
  const [notifications, setNotifications] = useState([]);
  const [storageReady, setStorageReady] = useState(false);
  const [newOrderToast, setNewOrderToast] = useState({ show: false, items: [] });

  useEffect(() => {
    try {
      const stored = localStorage.getItem('bakeflow_notifications');
      if (!stored) return;
      const parsed = JSON.parse(stored);
      if (Array.isArray(parsed)) setNotifications(parsed);
    } catch (e) {
      try {
        localStorage.removeItem('bakeflow_notifications');
      } catch (_) {}
    } finally {
      setStorageReady(true);
    }
  }, []);

  // Persist to localStorage whenever notifications change
  useEffect(() => {
    if (!storageReady) return;
    if (notifications.length === 0) {
      console.log('💾 Clearing localStorage (no notifications)');
      localStorage.removeItem('bakeflow_notifications');
    } else {
      console.log('💾 Saving', notifications.length, 'notifications to localStorage');
      localStorage.setItem('bakeflow_notifications', JSON.stringify(notifications));
    }
  }, [notifications, storageReady]);

  const addNotifications = useCallback((newOnes) => {
    if (!newOnes || newOnes.length === 0) return;
    console.log('🔔 Adding notifications:', newOnes);
    
    // Mark new notifications as unread and add timestamp
    const timestampedNotifications = newOnes.map(n => ({
      ...n,
      read: false,
      timestamp: n.timestamp || Date.now(),
      isRead: false,
      type: 'new_order'
    }));
    
    // Keep only last 20 notifications to prevent clutter
    setNotifications(prev => [...timestampedNotifications, ...prev].slice(0, 20));
    setNewOrderToast({ show: true, items: timestampedNotifications });
  }, []);

  const markAsRead = useCallback((notificationId) => {
    setNotifications(prev => 
      prev.map(n => n.id === notificationId ? { ...n, read: true, isRead: true, readAt: Date.now() } : n)
    );
  }, []);

  const markAllRead = useCallback(() => {
    const now = Date.now();
    setNotifications(prev => prev.map(n => ({ ...n, read: true, isRead: true, readAt: now })));
  }, []);

  const clearAll = useCallback(() => {
    setNotifications([]);
    try {
      localStorage.removeItem('bakeflow_notifications');
    } catch (_) {}
  }, []);

  const dismissToast = useCallback(() => {
    setNewOrderToast({ show: false, items: [] });
  }, []);

  // Auto-hide toast after 6s
  useEffect(() => {
    if (newOrderToast.show) {
      const t = setTimeout(dismissToast, 6000);
      return () => clearTimeout(t);
    }
  }, [newOrderToast.show, dismissToast]);

  const unreadCount = notifications.filter(n => !n.read && !n.isRead).length;
  const hasUnread = unreadCount > 0;

  return (
    <NotificationContext.Provider value={{
      notifications,
      newOrderToast,
      unreadCount,
      hasUnread,
      addNotifications,
      markAsRead,
      markAllRead,
      clearAll,
      dismissToast,
    }}>
      {children}
    </NotificationContext.Provider>
  );
}

export function useNotifications() {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotifications must be used within NotificationProvider');
  }
  return context;
}
