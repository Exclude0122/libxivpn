#include <jni.h>
#include <stdlib.h>
#include <android/log.h>

void libxivpn_log(char * msg) {
	__android_log_write(ANDROID_LOG_INFO, "libxivpn", msg);
	free(msg);
}

extern char* libxivpn_version();
extern char* libxivpn_start(char* config, int socksPort, int fd, char* logFile);
extern void libxivpn_stop();


JNIEXPORT jstring JNICALL Java_cn_gov_xivpn2_LibXivpn_xivpn_1version (JNIEnv * env, jclass clazz)
{
	char *v = libxivpn_version();
	jstring s = (*env)->NewStringUTF(env, v);
	free(v);
	return s;
}

JNIEXPORT jstring JNICALL Java_cn_gov_xivpn2_LibXivpn_xivpn_1start (JNIEnv * env, jclass clazz, jstring config, jint socksPort, jint fd, jstring logFile) {
	const char *cConfig = (*env)->GetStringUTFChars(env, config, 0);
	const char *cLogFile = (*env)->GetStringUTFChars(env, logFile, 0);

	char *ret = libxivpn_start(cConfig, (int)socksPort, (int)fd, cLogFile);

	(*env)->ReleaseStringUTFChars(env, config, cConfig);
	(*env)->ReleaseStringUTFChars(env, config, cLogFile);

	jstring ret_jstring = (*env)->NewStringUTF(env, ret);
	free(ret);

	return ret_jstring;
}

JNIEXPORT void JNICALL Java_cn_gov_xivpn2_LibXivpn_xivpn_1stop (JNIEnv * env, jclass clazz) {
	libxivpn_stop();
}
