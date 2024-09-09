import { Subject, Observable } from "rxjs";
import { filter } from "rxjs/operators";

const alertSubject = new Subject<AlertTyping>();
const defaultId = "default-alert";

export const alertService = {
  onAlert,
  success,
  error,
  info,
  warn,
  alert,
  clear,
};

export const AlertType = {
  Success: "success",
  Error: "danger",
  Info: "info",
  Warning: "warning",
};

// enable subscribing to alerts observable
function onAlert(id = defaultId): Observable<AlertTyping> {
  return alertSubject.asObservable().pipe(filter((x) => x.id === id));
}

// convenience methods
function success(message: string, options: any) {
  alert({ ...options, type: AlertType.Success, message });
}

function error(message: any, options: any) {
  alert({ ...options, type: AlertType.Error, message });
}

function info(message: any, options: any) {
  alert({ ...options, type: AlertType.Info, message });
}

function warn(message: any, options: any) {
  alert({ ...options, type: AlertType.Warning, message });
}

export type AlertTyping = {
  id: string;
  type?: string;
  message?: string;
  keepAfterRouteChange?: boolean;
  autoClose?: boolean;
  fade?: boolean;
};

// core alert method
function alert(alert: AlertTyping) {
  alert.id = alert.id || defaultId;
  alertSubject.next(alert);
}

// clear alerts
function clear(id = defaultId) {
  alertSubject.next({ id });
}
