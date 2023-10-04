'use client'

import './globals.css'
import '../node_modules/bootstrap/dist/css/bootstrap.css'

import { Inter } from 'next/font/google'
import Dashboard from './dashboard'

const inter = Inter({ subsets: ['latin'] })

export default function RootLayout({
  children
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" data-bs-theme="dark">
      <body className={inter.className}>
          <Dashboard>{children}</Dashboard>
      </body>
    </html>
  )
}
