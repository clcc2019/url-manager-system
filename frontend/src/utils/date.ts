export const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
};

export const formatDuration = (seconds: number): string => {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (days > 0) {
    return `${days}天${hours}小时`;
  } else if (hours > 0) {
    return `${hours}小时${minutes}分钟`;
  } else {
    return `${minutes}分钟`;
  }
};

export const getTimeUntilExpiry = (expireAt: string): string => {
  const now = new Date();
  const expiry = new Date(expireAt);
  const diffInSeconds = Math.floor((expiry.getTime() - now.getTime()) / 1000);
  
  if (diffInSeconds <= 0) {
    return '已过期';
  }
  
  return formatDuration(diffInSeconds);
};