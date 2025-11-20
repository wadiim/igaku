export function sendNotification(message: string) {
  if (!("Notification" in window)) {
    console.error("Desktop notifications are not supported");
  } else if (Notification.permission === "granted") {
    _create_notification(message);
  } else if (Notification.permission !== "denied") {
    Notification.requestPermission().then((permission) => {
      if (permission === "granted") {
        _create_notification(message)
      }
    });
  }
}

function _create_notification(message: string) {
  new Notification("Igaku", {
    body: message,
    icon: "/logo.svg",
  });
}
