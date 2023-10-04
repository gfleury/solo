'use client'
import * as apiClient from '../api-client'
import Loading from '../loading'

export default function Home() {
    const {data, isLoading} = apiClient.Get("/user/login")

    if (isLoading) return ( Loading() )
 
   return (<div>Hellow {data.firstName}!</div>)
}