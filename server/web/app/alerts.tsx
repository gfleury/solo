import React, { useState, useEffect } from 'react';
import { alertService, AlertTyping } from './services/alerts';
import Alert from 'react-bootstrap/Alert'

type AlertProps = {
    id?: string,
    fade?: boolean
};


export default function MyAlert({ id = 'default-alert', fade = true }: AlertProps) {
    const alertsArray: AlertTyping[] = []
    const [alerts, setAlerts] = useState(alertsArray);


    useEffect(() => {

        function removeAlert(alert: AlertTyping) {
            if (fade) {
                // fade out alert
                const alertWithFade = { ...alert, fade: true };
                setAlerts(alerts => alerts.map(x => x === alert ? alertWithFade : x));

                // remove alert after faded out
                setTimeout(() => {
                    setAlerts(alerts => alerts.filter(x => x !== alertWithFade));
                }, 250);
            } else {
                // remove alert
                setAlerts(alerts => alerts.filter(x => x !== alert));
            }
        }

        // subscribe to new alert notifications
        const subscription = alertService.onAlert(id)
            .subscribe(alert => {
                // clear alerts when an empty alert is received
                if (!alert.message) {
                    setAlerts(alerts => {
                        // filter out alerts without 'keepAfterRouteChange' flag
                        const filteredAlerts = alerts.filter(x => x.keepAfterRouteChange);

                        // remove 'keepAfterRouteChange' flag on the rest
                        filteredAlerts.forEach(x => delete x.keepAfterRouteChange);
                        return filteredAlerts;
                    });
                } else {
                    // add alert to array
                    setAlerts(alerts => ([...alerts, alert]));

                    // auto close alert if required
                    if (alert.autoClose) {
                        setTimeout(() => removeAlert(alert), 3000);
                    }
                }
            });


        // clean up function that runs when the component unmounts
        return () => {
            // unsubscribe & unlisten to avoid memory leaks
            subscription.unsubscribe();
        };
    }, [id, fade]);


    if (!alerts.length) return (<></>);

    return (
        alerts.map((alert: AlertTyping) =>
            <Alert key={alert.id} variant={alert.type} dismissible>{alert.message}</Alert>
        )
    );
}
