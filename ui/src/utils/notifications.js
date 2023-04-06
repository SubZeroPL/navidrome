export const sendNotification = (title, body = '', image = '') => {
  checkForNotificationPermission()
  new Notification(title, {
    body: body,
    icon: image,
    silent: false,
    tag: 'Navidrome',
  })
}

const checkForNotificationPermission = () => {
  return 'Notification' in window && Notification.permission === 'granted'
}
