export function statusColor(status) {
  switch (status) {
    case 'scheduled': return 'secondary';
    case 'pending': return 'warning';
    case 'preparing': return 'primary';
    case 'ready': return 'info';
    case 'delivered': return 'success';
    default: return 'secondary';
  }
}
