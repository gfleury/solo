import { Dispatch, SetStateAction, createContext } from 'react';
import { User } from './api-client/models';

export type SessionStore = {
    user?: User
}

type SessionContextType = {
    session?: SessionStore,
    setSession?: Dispatch<SetStateAction<SessionStore>>,
}

const sessionContext: SessionContextType = {}
export const SessionContext = createContext(sessionContext)
