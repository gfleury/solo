import Image from 'next/image';
import { GetUrlAPI } from './api-client';

export default function MyImage(props: any) {
    const originalSrc = props.src
    const newProps = {...props}
    newProps.src = GetUrlAPI() + "/image?url=" + originalSrc
    return (<Image {...newProps}/>)
}