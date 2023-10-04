import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import Card from 'react-bootstrap/Card';
import randomColor from "randomcolor";
import { Get } from '../api-client';

export default function Graph() {
    const metrics = Get("/account/17/info/insights")
    if (metrics.isLoading || metrics.error) return <></>
    console.log(metrics.data.merged_metrics)
    
    const merged_metrics = metrics.data.merged_metrics

    function getDataPoint(name: string) {
        return function(datapoint: any): number {
            return datapoint.values[name].value
        }
    }

    return (
        <>
            <Card className="p-2">
                <Card.Body>
                    <Card.Title>Insight Metrics</Card.Title>
                    <Card.Text>
                        <ResponsiveContainer width="100%" height={200}>
                            <LineChart
                                data={merged_metrics.values}
                                margin={{
                                    top: 5,
                                    right: 30,
                                    left: 20,
                                    bottom: 5,
                                }}
                            >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="end_time" />
                                <YAxis />
                                <Tooltip />
                                <Legend />
                                {
                                merged_metrics.names.map((name: string) =>
                                    {
                                        return <Line key={name} name={name.replaceAll("_", " ")} type="monotone" dataKey={getDataPoint(name)} strokeWidth={2} stroke={randomColor()} />;
                                    }
                                )
                                }
                            </LineChart>
                        </ResponsiveContainer>
                    </Card.Text>
                </Card.Body>
            </Card>
            <br />
        </>
    )
}
