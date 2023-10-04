/**
 * 
 * @export
 * @interface Account
 */
export interface Account {
    /**
     * 
     * @type {number}
     * @memberof Account
     */
    ID?: number;
    /**
     * 
     * @type {string}
     * @memberof Account
     */
    name: string;
    /**
     * 
     * @type {string}
     * @memberof Account
     */
    login: string;
    /**
     * 
     * @type {string}
     * @memberof Account
     */
    password: string;
    /**
     * 
     * @type {Provider}
     * @memberof Account
     */
    provider: Provider;
    /**
     * 
     * @type {Session}
     * @memberof Account
     */
    session?: Session;
    /**
     * 
     * @type {Array<Tag>}
     * @memberof Account
     */
    tags?: Array<Tag>;
    /**
     * account status
     * @type {string}
     * @memberof Account
     */
    status?: Account.StatusEnum;
    /**
     * 
     * @type {string}
     * @memberof Account
     */
    avatar?: string;
    twofa?: string;
}

/**
 * @export
 * @namespace Account
 */
export namespace Account {
    /**
     * @export
     * @enum {string}
     */
    export enum StatusEnum {
        Available = 'available',
        Pending = 'pending',
        Broken = 'broken'
    }
}
/**
 * 
 * @export
 * @interface LinkedUser
 */
export interface LinkedUser {
    /**
     * 
     * @type {User}
     * @memberof LinkedUser
     */
    user?: User;
    /**
     * account status
     * @type {string}
     * @memberof LinkedUser
     */
    permissions?: LinkedUser.PermissionsEnum;
}

/**
 * @export
 * @namespace LinkedUser
 */
export namespace LinkedUser {
    /**
     * @export
     * @enum {string}
     */
    export enum PermissionsEnum {
        Write = 'write',
        Read = 'read',
        Owner = 'owner'
    }
}
/**
 * 
 * @export
 * @interface Provider
 */
export interface Provider {
    /**
     * 
     * @type {number}
     * @memberof Provider
     */
    ID?: number;
    /**
     * 
     * @type {string}
     * @memberof Provider
     */
    name?: string;
    authenticationType?: string;
}
/**
 * 
 * @export
 * @interface Session
 */
export interface Session {
    /**
     * 
     * @type {number}
     * @memberof Session
     */
    ID?: number;
    /**
     * 
     * @type {number}
     * @memberof Session
     */
    accountID?: number;
    /**
     * 
     * @type {string}
     * @memberof Session
     */
    sessionData?: string;
}
/**
 * 
 * @export
 * @interface Tag
 */
export interface Tag {
    /**
     * 
     * @type {number}
     * @memberof Tag
     */
    ID?: number;
    /**
     * 
     * @type {string}
     * @memberof Tag
     */
    name?: string;
}
/**
 * 
 * @export
 * @interface User
 */
export interface User {
    /**
     * 
     * @type {number}
     * @memberof User
     */
    ID?: number;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    username?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    firstName?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    lastName?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    email?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    password?: string;
    /**
     * 
     * @type {string}
     * @memberof User
     */
    phone?: string;
    /**
     * User Status
     * @type {number}
     * @memberof User
     */
    userStatus?: number;
    /**
     * Accounts that belong to user
     * @type {Array<Account>}
     * @memberof User
     */
    accounts?: Array<Account>;
    /**
     * Users that are linked to this user
     * @type {Array<LinkedUser>}
     * @memberof User
     */
    linkedUsers?: Array<LinkedUser>;
    /**
     * User has confirmed email
     * @type {boolean}
     * @memberof User
     */
    verified?: boolean;
    avatar?: string;
}