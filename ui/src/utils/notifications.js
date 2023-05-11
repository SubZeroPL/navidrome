export const sendNotification = (title, body = '', image = '') => {
  if (!checkForNotificationPermission())
    return;
  new Notification(title, {
    body: body,
    icon: image,
    silent: true,
    tag: 'Navidrome',
  })
}

const checkForNotificationPermission = () => {
  return 'Notification' in window && Notification.permission === 'granted'
}
